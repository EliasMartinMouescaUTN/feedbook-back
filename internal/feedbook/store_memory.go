package feedbook

import (
	"errors"
	"sort"
	"sync"
	"time"
)

type Store struct {
	mu              sync.RWMutex
	avatarPresets   []AvatarPreset
	ownProfile      Profile
	publicProfile   Profile
	stats           Stats
	notifications   Notifications
	books           []Book
	exploreUsers    []ExploreUser
	readingProgress map[string]ReadingProgress
	reviews         map[string][]Review
	reviewLikes     map[string]map[string]bool
	authors         []Author
	followedAuthors map[string]bool
}

func NewMemoryStore() *Store {
	store := &Store{
		avatarPresets:   avatarPresets(),
		stats:           sampleStats(),
		notifications:   sampleNotifications(),
		books:           sampleBooks(),
		exploreUsers:    sampleExploreUsers(),
		readingProgress: sampleReadingProgress(),
		reviews:         sampleReviews(),
		reviewLikes:     map[string]map[string]bool{},
		followedAuthors: map[string]bool{},
	}
	store.ownProfile = sampleOwnProfile(store.avatarPresets)
	store.publicProfile = samplePublicProfile(store.avatarPresets)
	store.authors = sampleAuthors(store.books)
	return store
}

func (s *Store) AvatarPresets() []AvatarPreset {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]AvatarPreset(nil), s.avatarPresets...)
}

func (s *Store) OwnProfile() Profile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ownProfile
}

func (s *Store) SetOwnProfile(profile Profile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ownProfile = profile
}

func (s *Store) PublicProfile() Profile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.publicProfile
}

func (s *Store) Stats() Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stats
}

func (s *Store) Notifications() Notifications {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.notifications
}

func (s *Store) Books() []Book {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Book(nil), s.books...)
}

func (s *Store) BookByID(bookID string) (Book, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, book := range s.books {
		if book.ID == bookID {
			return book, true
		}
	}
	return Book{}, false
}

func (s *Store) ExploreUsers() []ExploreUser {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]ExploreUser(nil), s.exploreUsers...)
}

func (s *Store) ReadingProgress(bookID string) (ReadingProgress, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	progress, ok := s.readingProgress[bookID]
	return progress, ok
}

func (s *Store) SetReadingProgress(bookID string, currentPage int, totalPages int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	found := false
	for _, b := range s.books {
		if b.ID == bookID {
			found = true
			break
		}
	}
	if !found {
		return ErrNotFound
	}
	s.readingProgress[bookID] = ReadingProgress{
		BookID:      bookID,
		CurrentPage: currentPage,
		TotalPages:  totalPages,
		UpdatedAt:   time.Now().Format("02/01/2006"),
	}
	return nil
}

func (s *Store) Reviews(bookID string, page int, limit int) ([]Review, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	all := s.reviews[bookID]
	sorted := make([]Review, len(all))
	copy(sorted, all)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Likes != sorted[j].Likes {
			return sorted[i].Likes > sorted[j].Likes
		}
		return sorted[i].CreatedAt > sorted[j].CreatedAt
	})
	total := len(sorted)
	start := (page - 1) * limit
	if start >= total {
		return []Review{}, total
	}
	end := start + limit
	if end > total {
		end = total
	}
	return sorted[start:end], total
}

func (s *Store) SaveReview(bookID string, review Review) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	reviews := s.reviews[bookID]
	for i, r := range reviews {
		if r.UserID == review.UserID {
			reviews[i] = review
			s.reviews[bookID] = reviews
			return nil
		}
	}
	s.reviews[bookID] = append(reviews, review)
	return nil
}

func (s *Store) ToggleLike(userID string, reviewID string) (Review, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for bookID, reviews := range s.reviews {
		for i, r := range reviews {
			if r.ID != reviewID {
				continue
			}
			likes := s.reviewLikes[reviewID]
			if likes == nil {
				likes = map[string]bool{}
				s.reviewLikes[reviewID] = likes
			}
			if likes[userID] {
				delete(likes, userID)
			} else {
				likes[userID] = true
			}
			r.LikedBy = make([]string, 0, len(likes))
			for uid := range likes {
				r.LikedBy = append(r.LikedBy, uid)
			}
			sort.Strings(r.LikedBy)
			r.Likes = len(r.LikedBy)
			reviews[i] = r
			s.reviews[bookID] = reviews
			return r, nil
		}
	}
	return Review{}, ErrNotFound
}

func (s *Store) Authors() []Author {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Author(nil), s.authors...)
}

func (s *Store) AuthorByID(authorID string) (Author, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, author := range s.authors {
		if author.ID == authorID {
			return author, true
		}
	}
	return Author{}, false
}

func (s *Store) ToggleFollow(authorID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.followedAuthors[authorID] {
		delete(s.followedAuthors, authorID)
		return false
	}
	s.followedAuthors[authorID] = true
	return true
}

var ErrAlreadyInLibrary = errors.New("book already in library")

func (s *Store) AddBookToLibrary(bookID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var book Book
	for _, b := range s.books {
		if b.ID == bookID {
			book = b
			break
		}
	}
	if book.ID == "" {
		return ErrNotFound
	}
	for _, lb := range s.ownProfile.PublicLibrary {
		if lb.Title == book.Title {
			return ErrAlreadyInLibrary
		}
	}
	s.ownProfile.PublicLibrary = append(s.ownProfile.PublicLibrary, LibraryBook{
		ID:            book.ID,
		Title:         book.Title,
		CoverImageURL: book.CoverImageURL,
	})
	return nil
}

func (s *Store) RemoveBookFromLibrary(bookID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var book Book
	for _, b := range s.books {
		if b.ID == bookID {
			book = b
			break
		}
	}
	if book.ID == "" {
		return ErrNotFound
	}
	for i, lb := range s.ownProfile.PublicLibrary {
		if lb.ID == bookID {
			s.ownProfile.PublicLibrary = append(s.ownProfile.PublicLibrary[:i], s.ownProfile.PublicLibrary[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

func avatarPresets() []AvatarPreset {
	return []AvatarPreset{
		{ID: "vampire", TopColorHex: 0xFF382845, BottomColorHex: 0xFFBFA7CF, ImageURL: avatarURL("vampire")},
		{ID: "werewolf", TopColorHex: 0xFF4A3C32, BottomColorHex: 0xFFC8AE96, ImageURL: avatarURL("werewolf")},
		{ID: "witch", TopColorHex: 0xFF344B39, BottomColorHex: 0xFFC8D3B5, ImageURL: avatarURL("witch")},
		{ID: "wizard", TopColorHex: 0xFF29496B, BottomColorHex: 0xFFC5D5E8, ImageURL: avatarURL("wizard")},
		{ID: "harry_potter", TopColorHex: 0xFF6B2E2A, BottomColorHex: 0xFFE5C77F, ImageURL: avatarURL("harry-potter")},
		{ID: "astronaut", TopColorHex: 0xFF24364D, BottomColorHex: 0xFFCAD8E7, ImageURL: avatarURL("astronaut")},
		{ID: "grim_reaper", TopColorHex: 0xFF2B2B31, BottomColorHex: 0xFFB8BBC4, ImageURL: avatarURL("grim-reaper")},
		{ID: "fairy", TopColorHex: 0xFF5B4A80, BottomColorHex: 0xFFF0CCE9, ImageURL: avatarURL("fairy")},
		{ID: "pirate", TopColorHex: 0xFF5A3527, BottomColorHex: 0xFFE2C09A, ImageURL: avatarURL("pirate")},
		{ID: "princess", TopColorHex: 0xFF9A5C8D, BottomColorHex: 0xFFF2D8EB, ImageURL: avatarURL("princess")},
		{ID: "king", TopColorHex: 0xFF70511F, BottomColorHex: 0xFFF0D9A0, ImageURL: avatarURL("king")},
		{ID: "ghost", TopColorHex: 0xFF5B6775, BottomColorHex: 0xFFE6EBF0, ImageURL: avatarURL("ghost")},
	}
}

func sampleOwnProfile(presets []AvatarPreset) Profile {
	return Profile{
		Name:                   "Evelyn Vance",
		Handle:                 "@evelynv",
		Quote:                  "\"Reading is a conversation. All books talk. But a good book listens as well.\"",
		Avatar:                 Avatar{TopColorHex: 0xFF5B4A80, BottomColorHex: 0xFFF0CCE9, AvatarPresetID: "witch", PresetImageURL: avatarURL("witch")},
		AvailableAvatarPresets: presets,
		ReadingGoal:            &ReadingGoal{TargetPagesPerDay: 40, CurrentAveragePagesPerDay: 28},
		ReadingStreak: ReadingStreak{
			Days: 5,
			Week: []StreakDay{
				{Label: "M", FillFraction: 0.18},
				{Label: "T", FillFraction: 0.72, Completed: true},
				{Label: "W", FillFraction: 1, Completed: true},
				{Label: "T", FillFraction: 0.48, Completed: true},
				{Label: "F", FillFraction: 1, Completed: true},
				{Label: "S", FillFraction: 0.88, Completed: true},
				{Label: "S", IsToday: true},
			},
		},
		CurrentBook: CurrentBook{
			ID:            "1",
			Title:         "The Secret History",
			Author:        "Donna Tartt",
			Page:          248,
			TotalPages:    559,
			Progress:      0.44,
			CoverImageURL: coverURL("9781400031702"),
		},
		UpNextBooks: []QueuedBook{
			{Title: "Foucault's Pendulum", Author: "Umberto Eco", CoverImageURL: coverURL("9780156032971")},
			{Title: "The Shadow of the Wind", Author: "Carlos Ruiz Zafon", CoverImageURL: coverURL("9780143034902")},
			{Title: "If on a winter's night a traveler", Author: "Italo Calvino", CoverImageURL: coverURL("9780156439619")},
		},
		CompletedBooks: 142,
		ProfileStats: []ProfileStat{
			{Label: "Books read", Value: "142"},
			{Label: "This year", Value: "19"},
		},
		PublicLibrary: []LibraryBook{},
		FeaturedReviews: []FeaturedReview{
			{BookTitle: "The Secret History", Rating: 5, TimeAgo: "2d ago", Excerpt: "\"A novel built on obsession, elitism and silence. Tartt makes every scene feel both intimate and dangerous.\"", CoverImageURL: coverURL("9781400031702")},
			{BookTitle: "Beloved", Rating: 5, TimeAgo: "1w ago", Excerpt: "\"Morrison writes memory like weather. Every return to this novel feels heavier and more precise.\"", CoverImageURL: coverURL("9781400033416")},
		},
	}
}

func samplePublicProfile(presets []AvatarPreset) Profile {
	return Profile{
		Name:                   "Julian Thorne",
		Handle:                 "@julianthorne",
		Quote:                  "\"I collect stories that feel like half-remembered dreams and impossible cities.\"",
		Avatar:                 Avatar{TopColorHex: 0xFF5A3527, BottomColorHex: 0xFFE2C09A, AvatarPresetID: "pirate", PresetImageURL: avatarURL("pirate")},
		AvailableAvatarPresets: presets,
		ReadingStreak:          ReadingStreak{Week: []StreakDay{{Label: "M"}, {Label: "T"}, {Label: "W"}, {Label: "T"}, {Label: "F"}, {Label: "S"}, {Label: "S", IsToday: true}}},
		CurrentBook: CurrentBook{
			ID:            "2",
			Title:         "The Name of the Rose",
			Author:        "Umberto Eco",
			Page:          312,
			TotalPages:    512,
			Progress:      0.61,
			CoverImageURL: coverURL("9780156001311"),
		},
		UpNextBooks:    []QueuedBook{},
		CompletedBooks: 58,
		ProfileStats: []ProfileStat{
			{Label: "Reviews", Value: "128"},
			{Label: "Followers", Value: "2.4K"},
		},
		PublicLibrary: []LibraryBook{
			{ID: "one-hundred-years", Title: "One Hundred Years of Solitude", CoverImageURL: coverURL("9780060883287")},
			{ID: "shadow-wind", Title: "The Shadow of the Wind", CoverImageURL: coverURL("9780143034902")},
			{ID: "ficciones", Title: "Ficciones", CoverImageURL: coverURL("9780802130303")},
			{ID: "invisible-cities", Title: "Invisible Cities", CoverImageURL: coverURL("9780156453806")},
			{ID: "austerlitz", Title: "Austerlitz", CoverImageURL: coverURL("9780811216548")},
			{ID: "winters-night", Title: "If on a winter's night a traveler", CoverImageURL: coverURL("9780156439619")},
			{ID: "left-hand-darkness", Title: "The Left Hand of Darkness", CoverImageURL: coverURL("9780441478125")},
			{ID: "pedro-paramo", Title: "Pedro Paramo", CoverImageURL: coverURL("9780802133908")},
			{ID: "master-margarita", Title: "The Master and Margarita", CoverImageURL: coverURL("9780143108276")},
		},
		FeaturedReviews: []FeaturedReview{
			{BookTitle: "The Name of the Rose", Rating: 5, TimeAgo: "4h ago", Excerpt: "\"A profound meditation on destiny. The novel keeps its labyrinth open long after the final page.\"", CoverImageURL: coverURL("9780156001311")},
			{BookTitle: "Invisible Cities", Rating: 4, TimeAgo: "3d ago", Excerpt: "\"Calvino turns urban imagination into something light and exact. Every fragment expands after you finish it.\"", CoverImageURL: coverURL("9780156453806")},
			{BookTitle: "Austerlitz", Rating: 5, TimeAgo: "1w ago", Excerpt: "\"A quiet, relentless novel. Sebald makes memory feel architectural, fragile and impossible to escape.\"", CoverImageURL: coverURL("9780811216548")},
		},
	}
}

func sampleStats() Stats {
	return Stats{
		Title:         "Reading Ledger",
		Subtitle:      "A comprehensive overview of your literary engagement and year-to-date metrics.",
		Metrics:       []StatsMetric{{Label: "BOOKS READ", Value: "42"}, {Label: "TOTAL PAGES", Value: "12,450"}, {Label: "UNIQUE AUTHORS", Value: "38"}, {Label: "GENRES EXPLORED", Value: "12"}},
		HeatmapMonths: []string{"April", "May", "June"},
		HeatmapRows:   []string{"L", "M", "M", "J", "V", "S", "D"},
		HeatmapValues: [][]float32{{0.08, 0.12, 0.18, 0.15, 0.20, 0.28, 0.35, 0.55, 0.60, 0.72, 0.76, 0.68}, {0.05, 0.10, 0.16, 0.12, 0.22, 0.36, 0.45, 0.58, 0.62, 0.75, 0.82, 0.74}, {0.06, 0.09, 0.15, 0.18, 0.26, 0.33, 0.50, 0.57, 0.64, 0.78, 0.86, 0.80}, {0.04, 0.08, 0.14, 0.20, 0.29, 0.41, 0.47, 0.61, 0.67, 0.70, 0.78, 0.73}, {0.03, 0.10, 0.13, 0.22, 0.31, 0.38, 0.52, 0.56, 0.63, 0.71, 0.79, 0.76}, {0.02, 0.07, 0.12, 0.18, 0.24, 0.30, 0.43, 0.50, 0.58, 0.66, 0.74, 0.69}, {0.01, 0.05, 0.08, 0.14, 0.19, 0.26, 0.34, 0.41, 0.49, 0.57, 0.63, 0.60}},
		RadarSections: []RadarSection{
			{Mode: "Genre", Axes: []RadarAxis{{Label: "Adventure", Value: 0.46}, {Label: "Fantasy", Value: 0.68}, {Label: "Sci-Fi", Value: 0.54}, {Label: "Suspense", Value: 0.42}, {Label: "Horror", Value: 0.34}, {Label: "Romance", Value: 0.52}, {Label: "Drama", Value: 0.74}, {Label: "Mystery", Value: 0.63}}, Ranking: []RankingItem{{Rank: 1, Label: "Drama"}, {Rank: 2, Label: "Science Fiction"}, {Rank: 3, Label: "Fantasy"}, {Rank: 4, Label: "Mystery"}, {Rank: 5, Label: "Romance"}}},
			{Mode: "Author", Axes: []RadarAxis{{Label: "Asimov", Value: 0.72}, {Label: "Le Guin", Value: 0.58}, {Label: "Murakami", Value: 0.44}, {Label: "King", Value: 0.39}, {Label: "Austen", Value: 0.34}, {Label: "Doyle", Value: 0.49}, {Label: "Tolkien", Value: 0.81}, {Label: "Atwood", Value: 0.56}}, Ranking: []RankingItem{{Rank: 1, Label: "J.R.R. Tolkien"}, {Rank: 2, Label: "Isaac Asimov"}, {Rank: 3, Label: "Ursula K. Le Guin"}, {Rank: 4, Label: "Margaret Atwood"}, {Rank: 5, Label: "Arthur Conan Doyle"}}},
		},
	}
}

func sampleNotifications() Notifications {
	return Notifications{
		Title: "Activity and Notifications",
		Items: []NotificationEntry{
			{ID: "notif_1", Type: "liked_your_review", Timestamp: "HOY · 10:24", Actor: NotificationActor{Name: "Juan", AvatarTopColorHex: 0xFF35566F, AvatarBottomColorHex: 0xFFC8A988}, FallbackText: "A Juan le gustó tu reseña."},
			{ID: "notif_2", Type: "followed_you", Timestamp: "HOY · 07:12", Actor: NotificationActor{Name: "Sofía", AvatarTopColorHex: 0xFFB9CBE3, AvatarBottomColorHex: 0xFFE7EEF7}, FallbackText: "Sofía comenzó a seguirte."},
			{ID: "notif_3", Type: "reviewed_book", Timestamp: "AYER · 14:30", Actor: NotificationActor{Name: "Elena", AvatarTopColorHex: 0xFF534D61, AvatarBottomColorHex: 0xFFD9B89C}, Book: &NotificationBookSummary{Title: "El Laberinto de los Espíritus", Author: "CARLOS RUIZ ZAFÓN", CoverImageURL: coverURL("9788408163381")}, FallbackText: "Elena hizo una reseña sobre un libro."},
			{ID: "notif_4", Type: "started_reading", Timestamp: "AYER · 09:18", Actor: NotificationActor{Name: "Martina", AvatarTopColorHex: 0xFF6D7FA2, AvatarBottomColorHex: 0xFFDAB596}, Book: &NotificationBookSummary{Title: "The Left Hand of Darkness", Author: "URSULA K. LE GUIN", CoverImageURL: coverURL("9780441478125")}, FallbackText: "Martina empezó a leer un nuevo libro."},
			{ID: "notif_5", Type: "followed_you", Timestamp: "LUNES · 21:04", Actor: NotificationActor{Name: "Tomás", AvatarTopColorHex: 0xFF4E697F, AvatarBottomColorHex: 0xFFE6C7AA}, FallbackText: "Tomás comenzó a seguirte."},
			{ID: "notif_6", Type: "saved_your_book", Timestamp: "LUNES · 17:42", Actor: NotificationActor{Name: "Lucía", AvatarTopColorHex: 0xFF7A8B6A, AvatarBottomColorHex: 0xFFDCC6A7}, Book: &NotificationBookSummary{Title: "Beloved", Author: "TONI MORRISON", CoverImageURL: coverURL("9781400033416")}, FallbackText: "Lucía guardó uno de tus libros en su lista."},
			{ID: "notif_7", Type: "liked_your_review", Timestamp: "DOMINGO · 19:26", Actor: NotificationActor{Name: "Bruno", AvatarTopColorHex: 0xFF5A556A, AvatarBottomColorHex: 0xFFCDA58B}, FallbackText: "A Bruno le gustó tu reseña."},
			{ID: "notif_8", Type: "reviewed_book", Timestamp: "DOMINGO · 11:03", Actor: NotificationActor{Name: "Camila", AvatarTopColorHex: 0xFF7D6B8D, AvatarBottomColorHex: 0xFFE2C39F}, Book: &NotificationBookSummary{Title: "Piranesi", Author: "SUSANNA CLARKE", CoverImageURL: coverURL("9781635575637")}, FallbackText: "Camila hizo una reseña sobre un libro."},
			{ID: "notif_9", Type: "followed_you", Timestamp: "SÁBADO · 16:58", Actor: NotificationActor{Name: "Nicolás", AvatarTopColorHex: 0xFF4D6B73, AvatarBottomColorHex: 0xFFD3B08C}, FallbackText: "Nicolás comenzó a seguirte."},
			{ID: "notif_10", Type: "quote_liked", Timestamp: "SÁBADO · 08:41", Actor: NotificationActor{Name: "Irene", AvatarTopColorHex: 0xFF607D8B, AvatarBottomColorHex: 0xFFE5CDB4}, FallbackText: "A Irene le gustó tu cita destacada de \"The Waves\"."},
		},
	}
}

func sampleBooks() []Book {
	return []Book{
		{ID: "1", Title: "The Secret History", Author: "Donna Tartt", CoverImageURL: coverURL("9781400031702"), Genre: "Fiction", Description: "A group of classics students at a small Vermont college become entangled in a murder.", Pages: 559, Language: "English", Published: "27/05/1987", ISBN: "987698762"},
		{ID: "2", Title: "The Name of the Rose", Author: "Umberto Eco", CoverImageURL: coverURL("9780156001311"), Genre: "Mystery", Description: "A medieval monk investigates a series of mysterious deaths in an Italian abbey.", Pages: 242, Language: "English", Published: "27/05/1927", ISBN: "987618762"},
		{ID: "3", Title: "Beloved", Author: "Toni Morrison", CoverImageURL: coverURL("9781400033416"), Genre: "Fiction", Description: "A former enslaved woman is haunted by the ghost of her daughter.", Pages: 559, Language: "English", Published: "27/05/1986", ISBN: "987618763"},
	}
}

func sampleExploreUsers() []ExploreUser {
	return []ExploreUser{
		{ID: "user_1", Name: "Evelyn Vance", Handle: "@evelynv", Bio: "Reader of gothic fiction, essays and books that leave a trace.", AvatarImageURL: avatarURL("witch"), AvatarTopColorHex: 0xFF5B4A80, AvatarBottomColorHex: 0xFFF0CCE9, FollowersLabel: "2.1K seguidores", BooksReadLabel: "142 libros"},
		{ID: "user_2", Name: "Julian Thorne", Handle: "@julianthorne", Bio: "Collects impossible cities, labyrinths and dreamlike novels.", AvatarImageURL: avatarURL("pirate"), AvatarTopColorHex: 0xFF5A3527, AvatarBottomColorHex: 0xFFE2C09A, FollowersLabel: "2.4K seguidores", BooksReadLabel: "58 libros"},
		{ID: "user_3", Name: "Mina Solberg", Handle: "@minareads", Bio: "Nordic noir, quiet classics and deeply annotated rereads.", AvatarImageURL: avatarURL("ghost"), AvatarTopColorHex: 0xFF5B6775, AvatarBottomColorHex: 0xFFE6EBF0, FollowersLabel: "980 seguidores", BooksReadLabel: "87 libros"},
	}
}

func sampleReadingProgress() map[string]ReadingProgress {
	return map[string]ReadingProgress{
		"1": {BookID: "1", CurrentPage: 248, TotalPages: 559, UpdatedAt: "11/05/2026"},
		"2": {BookID: "2", CurrentPage: 312, TotalPages: 512, UpdatedAt: "11/05/2026"},
	}
}

func sampleReviews() map[string][]Review {
	return map[string][]Review{
		"1": {
			{ID: "r1", UserID: "me", ReviewerName: "Evelyn Vance", ReviewerAvatar: avatarURL("witch"), Rating: 5, Text: "A novel built on obsession, elitism and silence. Tartt makes every scene feel both intimate and dangerous.", Likes: 42, LikedBy: []string{}, CreatedAt: "2d ago"},
			{ID: "r2", UserID: "", ReviewerName: "Julian Thorne", ReviewerAvatar: avatarURL("pirate"), Rating: 4, Text: "Dense and rewarding. Every page pulls you deeper into its dark academia world.", Likes: 18, LikedBy: []string{}, CreatedAt: "1w ago"},
		},
		"2": {
			{ID: "r3", UserID: "", ReviewerName: "Julian Thorne", ReviewerAvatar: avatarURL("pirate"), Rating: 5, Text: "A profound meditation on destiny. The novel keeps its labyrinth open long after the final page.", Likes: 31, LikedBy: []string{}, CreatedAt: "4h ago"},
		},
		"3": {
			{ID: "r4", UserID: "", ReviewerName: "Evelyn Vance", ReviewerAvatar: avatarURL("witch"), Rating: 5, Text: "Morrison writes memory like weather. Every return to this novel feels heavier and more precise.", Likes: 67, LikedBy: []string{}, CreatedAt: "1w ago"},
		},
	}
}

func sampleAuthors(books []Book) []Author {
	return []Author{
		{ID: "a1", Name: "Donna Tartt", BirthYear: 1963, Nationality: "American", Description: "Pulitzer Prize-winning author known for her intricate literary fiction.", Biography: "Donna Tartt was born in Greenwood, Mississippi in 1963. She studied at the University of Mississippi and Bennington College, where she began writing her debut novel. Her first book, The Secret History, was published in 1992 to widespread acclaim. Known for her meticulous prose and infrequent output, she spent a decade on each of her novels. Her third novel, The Goldfinch, won the Pulitzer Prize for Fiction in 2014.", Books: filterBooks(books, "1"), Followers: 14200},
		{ID: "a2", Name: "Umberto Eco", BirthYear: 1932, DeathYear: intPtr(2016), Nationality: "Italian", Description: "Philosopher, semiotician and novelist renowned for his erudite fiction.", Biography: "Umberto Eco was born in Alessandria, Italy in 1932. A professor of semiotics at the University of Bologna, he became one of Italy's most celebrated intellectuals. His debut novel, The Name of the Rose, published in 1980, became an international bestseller and established him as a major literary figure. His works blend medieval history, philosophy, and literary theory into dense, rewarding narratives.", Books: filterBooks(books, "2"), Followers: 21500},
		{ID: "a3", Name: "Toni Morrison", BirthYear: 1931, DeathYear: intPtr(2019), Nationality: "American", Description: "Nobel Prize-winning author whose work explores the African American experience.", Biography: "Toni Morrison was born Chloe Ardelia Wofford in Lorain, Ohio in 1931. She studied at Howard University and Cornell, and worked as an editor at Random House before becoming a celebrated novelist. Her novel Beloved, published in 1987, won the Pulitzer Prize and later the Nobel Prize in Literature in 1993. Her prose, lyrical and unflinching, redefined American literature.", Books: filterBooks(books, "3"), Followers: 38900},
	}
}

func coverURL(isbn string) string {
	return "https://covers.openlibrary.org/b/isbn/" + isbn + "-L.jpg"
}

func avatarURL(seed string) string {
	return "https://api.dicebear.com/9.x/adventurer/png?seed=" + seed + "&size=128"
}

func filterBooks(books []Book, ids ...string) []Book {
	allowed := map[string]bool{}
	for _, id := range ids {
		allowed[id] = true
	}
	filtered := make([]Book, 0, len(ids))
	for _, book := range books {
		if allowed[book.ID] {
			filtered = append(filtered, book)
		}
	}
	return filtered
}

func intPtr(v int) *int {
	return &v
}
