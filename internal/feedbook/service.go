package feedbook

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

var ErrNotFound = errors.New("not found")
var ErrInvalidInput = errors.New("invalid input")

type Service struct {
	store Storer
}

func NewService(store Storer) *Service {
	return &Service{store: store}
}

func (s *Service) GetBooks() []Book { return s.store.Books() }

func (s *Service) GetBookByID(bookID string) (Book, error) {
	book, ok := s.store.BookByID(bookID)
	if !ok {
		return Book{}, ErrNotFound
	}
	return book, nil
}

func (s *Service) GetExploreUsers() []ExploreUser { return s.store.ExploreUsers() }

func (s *Service) GetReadingProgress(bookID string) (*ReadingProgress, error) {
	progress, ok := s.store.ReadingProgress(bookID)
	if !ok {
		return nil, ErrNotFound
	}
	return &progress, nil
}

func (s *Service) SaveReadingProgress(bookID string, currentPage int) (*ReadingProgress, error) {
	book, err := s.GetBookByID(bookID)
	if err != nil {
		return nil, err
	}
	if currentPage < 0 || currentPage > book.Pages {
		return nil, ErrInvalidInput
	}
	if err := s.store.SetReadingProgress(bookID, currentPage, book.Pages); err != nil {
		return nil, err
	}
	progress, ok := s.store.ReadingProgress(bookID)
	if !ok {
		return nil, ErrNotFound
	}
	return &progress, nil
}

func (s *Service) GetReviews(bookID string, page int, limit int) ([]Review, int) {
	return s.store.Reviews(bookID, page, limit)
}

func (s *Service) SaveReview(bookID string, rating float32, text string) (Review, error) {
	if _, err := s.GetBookByID(bookID); err != nil {
		return Review{}, err
	}
	if rating < 0 || rating > 5 {
		return Review{}, ErrInvalidInput
	}
	if strings.TrimSpace(text) == "" {
		return Review{}, ErrInvalidInput
	}
	profile := s.store.OwnProfile()
	review := Review{
		ID:           bookID + "-user-review",
		UserID:       "me",
		ReviewerName: profile.Name,
		Rating:       rating,
		Text:         text,
		Likes:        0,
		LikedBy:      []string{},
		CreatedAt:    time.Now().Format("Jan 02, 2006"),
	}
	if err := s.store.SaveReview(bookID, review); err != nil {
		return Review{}, err
	}
	return review, nil
}

func (s *Service) ToggleLike(bookID string, reviewID string) (Review, error) {
	return s.store.ToggleLike("me", reviewID)
}

func (s *Service) GetAuthors() []Author { return s.store.Authors() }

func (s *Service) GetAuthorByID(authorID string) (Author, error) {
	author, ok := s.store.AuthorByID(authorID)
	if !ok {
		return Author{}, ErrNotFound
	}
	return author, nil
}

func (s *Service) ToggleAuthorFollow(authorID string) error {
	if _, ok := s.store.AuthorByID(authorID); !ok {
		return ErrNotFound
	}
	s.store.ToggleFollow(authorID)
	return nil
}

func (s *Service) GetOwnProfile() Profile { return s.store.OwnProfile() }

func (s *Service) GetOwnPublicPreview() Profile {
	profile := s.store.OwnProfile()
	profile.ProfileStats = []ProfileStat{
		{Label: "Books read", Value: strconv.Itoa(profile.CompletedBooks)},
		{Label: "Daily goal", Value: dailyGoalLabel(profile.ReadingGoal)},
	}
	return profile
}

func (s *Service) GetPublicProfile() Profile { return s.store.PublicProfile() }

func (s *Service) GetHome() Home {
	profile := s.store.OwnProfile()
	return Home{
		TrendingTitle: "Trending Now",
		Avatar:        profile.Avatar,
		FeaturedBook:  HomeFeaturedBook{Label: "FEATURED", Title: "The Midnight Library", Author: "Matt Haig", CoverImageURL: "https://images.unsplash.com/photo-1512820790803-83ca734da794?auto=format&fit=crop&w=1200&q=80"},
		RankedBooks: []HomeRankedBook{
			{RankLabel: "01", Title: "Circe", Author: "Madeline Miller", CoverImageURL: coverURL("9780316556323")},
			{RankLabel: "02", Title: "Piranesi", Author: "Susanna Clarke", CoverImageURL: coverURL("9781635575637")},
			{RankLabel: "03", Title: "Project Hail Mary", Author: "Andy Weir", CoverImageURL: coverURL("9780593135204")},
		},
		ReadingRooms: []HomeReadingRoom{
			{HostName: "Eleanor", HostImageURL: avatarURL("eleanor"), Title: "Magical Realism Book Club", ReaderCountLabel: "1.2k readers"},
			{HostName: "James", HostImageURL: avatarURL("james"), Title: "20th Century Classics", ReaderCountLabel: "850 readers"},
		},
		Curators: []HomeCurator{
			{Name: "Dr. Aris Thorne", Focus: "Historical Non-Fiction Focus", ImageURL: avatarURL("aris-thorne")},
			{Name: "Lila Vance", Focus: "Contemporary Lit & Essays", ImageURL: avatarURL("lila-vance")},
		},
	}
}

func (s *Service) GetOwnLibrary() Library {
	profile := s.store.OwnProfile()

	var currentBook CurrentBook
	var highestProgress float32 = -1

	for _, lb := range profile.PublicLibrary {
		progress, ok := s.store.ReadingProgress(lb.ID)
		if !ok {
			continue
		}
		ratio := float32(progress.CurrentPage) / float32(progress.TotalPages)
		if ratio >= 1.0 {
			continue
		}
		if ratio > highestProgress {
			highestProgress = ratio
			book, _ := s.store.BookByID(lb.ID)
			currentBook = CurrentBook{
				ID:            lb.ID,
				Title:         lb.Title,
				Author:        book.Author,
				Page:          progress.CurrentPage,
				TotalPages:    progress.TotalPages,
				Progress:      ratio,
				CoverImageURL: lb.CoverImageURL,
			}
		}
	}

	return Library{
		Title:          "My Library",
		Subtitle:       "Your personal collection, current read, and completed shelf.",
		Avatar:         profile.Avatar,
		CurrentBook:    currentBook,
		ReadingBooks:   profile.PublicLibrary,
		ShelfBooks:     profile.PublicLibrary,
		CompletedBooks: profile.CompletedBooks,
		ReadHistory:    []ReadBook{},
	}
}

func (s *Service) AddBookToLibrary(bookID string) error {
	if _, err := s.GetBookByID(bookID); err != nil {
		return err
	}
	return s.store.AddBookToLibrary(bookID)
}

func (s *Service) RemoveBookFromLibrary(bookID string) error {
	if _, err := s.GetBookByID(bookID); err != nil {
		return err
	}
	return s.store.RemoveBookFromLibrary(bookID)
}

func (s *Service) UpdateOwnProfile(request UpdateProfileRequest) (Profile, error) {
	if strings.TrimSpace(request.Name) == "" || strings.TrimSpace(request.Handle) == "" || strings.TrimSpace(request.Quote) == "" {
		return Profile{}, ErrInvalidInput
	}
	if request.TargetPagesPerDay != nil && *request.TargetPagesPerDay <= 0 {
		return Profile{}, ErrInvalidInput
	}
	profile := s.store.OwnProfile()
	profile.Name = request.Name
	profile.Handle = request.Handle
	profile.Quote = request.Quote
	profile.Avatar.TopColorHex = request.AvatarTopColorHex
	profile.Avatar.BottomColorHex = request.AvatarBottomColorHex
	if request.AvatarPresetID != nil {
		profile.Avatar.AvatarPresetID = *request.AvatarPresetID
		profile.Avatar.PresetImageURL = ""
		for _, preset := range s.store.AvatarPresets() {
			if preset.ID == *request.AvatarPresetID {
				profile.Avatar.PresetImageURL = preset.ImageURL
				break
			}
		}
	}
	if request.AvatarImageURI != nil {
		profile.Avatar.ImageURI = *request.AvatarImageURI
	}
	if request.TargetPagesPerDay != nil {
		currentAverage := int(float32(*request.TargetPagesPerDay) * 0.7)
		if profile.ReadingGoal != nil {
			currentAverage = profile.ReadingGoal.CurrentAveragePagesPerDay
		}
		profile.ReadingGoal = &ReadingGoal{TargetPagesPerDay: *request.TargetPagesPerDay, CurrentAveragePagesPerDay: currentAverage}
	}
	s.store.SetOwnProfile(profile)
	return profile, nil
}

func (s *Service) GetStats() Stats { return s.store.Stats() }

func (s *Service) GetNotifications() Notifications { return s.store.Notifications() }

func dailyGoalLabel(goal *ReadingGoal) string {
	if goal == nil {
		return "None"
	}
	return strconv.Itoa(goal.TargetPagesPerDay) + " pgs"
}
