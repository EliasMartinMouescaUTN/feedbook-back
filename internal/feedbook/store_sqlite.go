package feedbook

import (
	"encoding/json"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SQLiteStore struct {
	db *gorm.DB
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&AccountModel{},
		&ProfileModel{},
		&AvatarModel{},
		&ReadingGoalModel{},
		&ReadingStreakModel{},
		&StreakDayModel{},
		&CurrentBookModel{},
		&QueuedBookModel{},
		&ProfileStatModel{},
		&PublicLibraryModel{},
		&FeaturedReviewModel{},
		&StatsModel{},
		&StatsMetricModel{},
		&RadarSectionModel{},
		&RadarAxisModel{},
		&RankingItemModel{},
		&HeatmapEntryModel{},
		&BookModel{},
		&ReviewModel{},
		&ReadingProgressModel{},
		&AuthorModel{},
		&ExploreUserModel{},
		&AvatarPresetModel{},
		&NotificationEntryModel{},
		&FollowedAuthorModel{},
		&ReviewLikeModel{},
	); err != nil {
		return nil, err
	}

	store := &SQLiteStore{db: db}
	store.ensureDefaultAccount()

	var count int64
	db.Model(&BookModel{}).Count(&count)
	if count == 0 {
		store.seed()
	}

	return store, nil
}

func (s *SQLiteStore) ensureDefaultAccount() {
	_, _ = s.CreateAccount("demo", "demo")
}

func (s *SQLiteStore) CreateAccount(username string, password string) (bool, error) {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	if username == "" || password == "" {
		return false, nil
	}

	var count int64
	if err := s.db.Model(&AccountModel{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, err
	}
	if count > 0 {
		return false, nil
	}

	if err := s.db.Create(&AccountModel{Username: username, Password: password}).Error; err != nil {
		return false, err
	}
	return true, nil
}

func (s *SQLiteStore) AccountPassword(username string) (string, bool, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return "", false, nil
	}

	var account AccountModel
	if err := s.db.First(&account, "username = ?", username).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", false, nil
		}
		return "", false, err
	}
	return account.Password, true, nil
}

func (s *SQLiteStore) seed() {
	presets := avatarPresets()
	for _, p := range presets {
		s.db.Create(&AvatarPresetModel{
			ID: p.ID, TopColorHex: p.TopColorHex,
			BottomColorHex: p.BottomColorHex, ImageURL: p.ImageURL,
		})
	}

	ownProfile := sampleOwnProfile(presets)
	s.seedProfile("own", ownProfile)

	publicProfile := samplePublicProfile(presets)
	s.seedProfile("public", publicProfile)

	s.seedStats("own", sampleStats())

	for _, b := range sampleBooks() {
		s.db.Create(&BookModel{
			ID: b.ID, Title: b.Title, Author: b.Author,
			Description: b.Description, CoverImageURL: b.CoverImageURL,
			Pages: b.Pages, ISBN: b.ISBN, Genre: b.Genre,
			Language: b.Language, Published: b.Published,
		})
	}

	for _, u := range sampleExploreUsers() {
		s.db.Create(&ExploreUserModel{
			ID: u.ID, Name: u.Name, Handle: u.Handle, Bio: u.Bio,
			AvatarImageURL: u.AvatarImageURL, AvatarTopColorHex: u.AvatarTopColorHex,
			AvatarBottomColorHex: u.AvatarBottomColorHex,
			FollowersLabel:       u.FollowersLabel, BooksReadLabel: u.BooksReadLabel,
		})
	}

	for bookID, progress := range sampleReadingProgress() {
		s.db.Create(&ReadingProgressModel{
			BookID: bookID, CurrentPage: progress.CurrentPage,
			TotalPages: progress.TotalPages, UpdatedAt: progress.UpdatedAt,
		})
	}

	for bookID, reviews := range sampleReviews() {
		for _, r := range reviews {
			likedByBytes, _ := json.Marshal(r.LikedBy)
			s.db.Create(&ReviewModel{
				ID: r.ID, BookID: bookID, UserID: r.UserID,
				ReviewerName: r.ReviewerName, ReviewerAvatar: r.ReviewerAvatar,
				Rating: r.Rating, Text: r.Text,
				Likes: r.Likes, LikedBy: string(likedByBytes), CreatedAt: r.CreatedAt,
			})
		}
	}

	books := sampleBooks()
	authors := sampleAuthors(books)
	for _, a := range authors {
		authorModel := AuthorModel{
			ID: a.ID, Name: a.Name, BirthYear: a.BirthYear,
			DeathYear: a.DeathYear, Nationality: a.Nationality,
			Description: a.Description, Biography: a.Biography,
			ImageURL: a.ImageURL, Followers: a.Followers,
		}
		bookIDs := make([]string, len(a.Books))
		for i, b := range a.Books {
			bookIDs[i] = b.ID
		}
		var authorBooks []BookModel
		s.db.Where("id IN ?", bookIDs).Find(&authorBooks)
		authorModel.Books = authorBooks
		s.db.Create(&authorModel)
	}

	notif := sampleNotifications()
	for _, n := range notif.Items {
		model := NotificationEntryModel{
			ID: n.ID, Type: n.Type, Timestamp: n.Timestamp,
			ActorName: n.Actor.Name, ActorTopHex: n.Actor.AvatarTopColorHex,
			ActorBottomHex: n.Actor.AvatarBottomColorHex,
			FallbackText:   n.FallbackText,
		}
		if n.Book != nil {
			title := n.Book.Title
			author := n.Book.Author
			url := n.Book.CoverImageURL
			model.BookTitle = &title
			model.BookAuthor = &author
			model.BookCoverURL = &url
		}
		s.db.Create(&model)
	}
}

func (s *SQLiteStore) seedProfile(profileType string, p Profile) {
	profileModel := ProfileModel{
		Type: profileType, Name: p.Name, Handle: p.Handle,
		Quote: p.Quote, CompletedBooks: p.CompletedBooks,
	}

	s.db.Create(&profileModel)

	s.db.Create(&AvatarModel{
		ProfileID: profileModel.ID, TopColorHex: p.Avatar.TopColorHex,
		BottomColorHex: p.Avatar.BottomColorHex,
		AvatarPresetID: p.Avatar.AvatarPresetID,
		PresetImageURL: p.Avatar.PresetImageURL,
		ImageURI:       p.Avatar.ImageURI,
	})

	if p.ReadingGoal != nil {
		s.db.Create(&ReadingGoalModel{
			ProfileID:                 profileModel.ID,
			TargetPagesPerDay:         p.ReadingGoal.TargetPagesPerDay,
			CurrentAveragePagesPerDay: p.ReadingGoal.CurrentAveragePagesPerDay,
		})
	}

	streakModel := ReadingStreakModel{
		ProfileID: profileModel.ID, DaysCount: p.ReadingStreak.Days,
	}
	s.db.Create(&streakModel)
	for _, d := range p.ReadingStreak.Week {
		s.db.Create(&StreakDayModel{
			ReadingStreakID: streakModel.ID, Label: d.Label,
			FillFraction: d.FillFraction, IsToday: d.IsToday, Completed: d.Completed,
		})
	}

	s.db.Create(&CurrentBookModel{
		ProfileID: profileModel.ID, BookID: p.CurrentBook.ID,
		Title: p.CurrentBook.Title, Author: p.CurrentBook.Author,
		Page: p.CurrentBook.Page, TotalPages: p.CurrentBook.TotalPages,
		Progress:      p.CurrentBook.Progress,
		CoverImageURL: p.CurrentBook.CoverImageURL,
	})

	for _, q := range p.UpNextBooks {
		s.db.Create(&QueuedBookModel{
			ProfileID: profileModel.ID, Title: q.Title,
			Author: q.Author, CoverImageURL: q.CoverImageURL,
		})
	}

	for _, ps := range p.ProfileStats {
		s.db.Create(&ProfileStatModel{
			ProfileID: profileModel.ID, Label: ps.Label, Value: ps.Value,
		})
	}

	for _, lb := range p.PublicLibrary {
		s.db.Create(&PublicLibraryModel{
			ProfileID: profileModel.ID, BookID: lb.ID,
			Title: lb.Title, CoverImageURL: lb.CoverImageURL,
		})
	}

	for _, fr := range p.FeaturedReviews {
		s.db.Create(&FeaturedReviewModel{
			ProfileID: profileModel.ID, BookTitle: fr.BookTitle,
			Rating: fr.Rating, TimeAgo: fr.TimeAgo,
			Excerpt: fr.Excerpt, CoverImageURL: fr.CoverImageURL,
		})
	}
}

func (s *SQLiteStore) seedStats(profileType string, st Stats) {
	var profile ProfileModel
	s.db.Where("type = ?", profileType).First(&profile)

	statsModel := StatsModel{
		ProfileID: profile.ID, Title: st.Title, Subtitle: st.Subtitle,
		HeatmapMonths: strings.Join(st.HeatmapMonths, "\n"),
		HeatmapRows:   strings.Join(st.HeatmapRows, "\n"),
	}
	s.db.Create(&statsModel)

	for _, m := range st.Metrics {
		s.db.Create(&StatsMetricModel{
			StatsID: statsModel.ID, Label: m.Label, Value: m.Value,
		})
	}

	for _, rs := range st.RadarSections {
		radarModel := RadarSectionModel{
			StatsID: statsModel.ID, Mode: rs.Mode,
		}
		s.db.Create(&radarModel)
		for _, ax := range rs.Axes {
			s.db.Create(&RadarAxisModel{
				RadarSectionID: radarModel.ID,
				Label:          ax.Label, Value: ax.Value,
			})
		}
		for _, rk := range rs.Ranking {
			s.db.Create(&RankingItemModel{
				RadarSectionID: radarModel.ID,
				Rank:           rk.Rank, Label: rk.Label,
			})
		}
	}

	for rowIdx, row := range st.HeatmapValues {
		for colIdx, val := range row {
			s.db.Create(&HeatmapEntryModel{
				StatsID: statsModel.ID, RowIndex: rowIdx,
				ColIndex: colIdx, Value: val,
			})
		}
	}
}

func (s *SQLiteStore) profileModelToDTO(m ProfileModel) Profile {
	var presets []AvatarPresetModel
	s.db.Find(&presets)

	avatarPresetDTOs := make([]AvatarPreset, len(presets))
	for i, p := range presets {
		avatarPresetDTOs[i] = AvatarPreset{
			ID: p.ID, TopColorHex: p.TopColorHex,
			BottomColorHex: p.BottomColorHex, ImageURL: p.ImageURL,
		}
	}

	p := Profile{
		Name:                   m.Name,
		Handle:                 m.Handle,
		Quote:                  m.Quote,
		CompletedBooks:         m.CompletedBooks,
		AvailableAvatarPresets: avatarPresetDTOs,
	}

	var avatar AvatarModel
	s.db.Where("profile_id = ?", m.ID).First(&avatar)
	p.Avatar = Avatar{
		TopColorHex:    avatar.TopColorHex,
		BottomColorHex: avatar.BottomColorHex,
		AvatarPresetID: avatar.AvatarPresetID,
		PresetImageURL: avatar.PresetImageURL,
		ImageURI:       avatar.ImageURI,
	}

	var goal ReadingGoalModel
	if err := s.db.Where("profile_id = ?", m.ID).First(&goal).Error; err == nil {
		p.ReadingGoal = &ReadingGoal{
			TargetPagesPerDay:         goal.TargetPagesPerDay,
			CurrentAveragePagesPerDay: goal.CurrentAveragePagesPerDay,
		}
	}

	var streak ReadingStreakModel
	if err := s.db.Where("profile_id = ?", m.ID).First(&streak).Error; err == nil {
		var days []StreakDayModel
		s.db.Where("reading_streak_id = ?", streak.ID).Find(&days)
		p.ReadingStreak.Days = streak.DaysCount
		p.ReadingStreak.Week = make([]StreakDay, len(days))
		for i, d := range days {
			p.ReadingStreak.Week[i] = StreakDay{
				Label: d.Label, FillFraction: d.FillFraction,
				IsToday: d.IsToday, Completed: d.Completed,
			}
		}
	}

	var cb CurrentBookModel
	if err := s.db.Where("profile_id = ?", m.ID).First(&cb).Error; err == nil {
		p.CurrentBook = CurrentBook{
			ID: cb.BookID, Title: cb.Title, Author: cb.Author,
			Page: cb.Page, TotalPages: cb.TotalPages,
			Progress: cb.Progress, CoverImageURL: cb.CoverImageURL,
		}
	}

	var queued []QueuedBookModel
	s.db.Where("profile_id = ?", m.ID).Find(&queued)
	p.UpNextBooks = make([]QueuedBook, len(queued))
	for i, q := range queued {
		p.UpNextBooks[i] = QueuedBook{
			Title: q.Title, Author: q.Author, CoverImageURL: q.CoverImageURL,
		}
	}

	var profileStats []ProfileStatModel
	s.db.Where("profile_id = ?", m.ID).Find(&profileStats)
	p.ProfileStats = make([]ProfileStat, len(profileStats))
	for i, ps := range profileStats {
		p.ProfileStats[i] = ProfileStat{Label: ps.Label, Value: ps.Value}
	}

	var library []PublicLibraryModel
	s.db.Where("profile_id = ?", m.ID).Find(&library)
	p.PublicLibrary = make([]LibraryBook, len(library))
	for i, lb := range library {
		p.PublicLibrary[i] = LibraryBook{
			ID: lb.BookID, Title: lb.Title, CoverImageURL: lb.CoverImageURL,
		}
	}

	var featured []FeaturedReviewModel
	s.db.Where("profile_id = ?", m.ID).Find(&featured)
	p.FeaturedReviews = make([]FeaturedReview, len(featured))
	for i, fr := range featured {
		p.FeaturedReviews[i] = FeaturedReview{
			BookTitle: fr.BookTitle, Rating: fr.Rating, TimeAgo: fr.TimeAgo,
			Excerpt: fr.Excerpt, CoverImageURL: fr.CoverImageURL,
		}
	}

	return p
}

func (s *SQLiteStore) statsModelToDTO(m StatsModel) Stats {
	var metrics []StatsMetricModel
	s.db.Where("stats_id = ?", m.ID).Find(&metrics)
	metricDTOs := make([]StatsMetric, len(metrics))
	for i, met := range metrics {
		metricDTOs[i] = StatsMetric{Label: met.Label, Value: met.Value}
	}

	var sections []RadarSectionModel
	s.db.Where("stats_id = ?", m.ID).Find(&sections)
	sectionDTOs := make([]RadarSection, len(sections))
	for i, rs := range sections {
		var axes []RadarAxisModel
		s.db.Where("radar_section_id = ?", rs.ID).Find(&axes)
		axisDTOs := make([]RadarAxis, len(axes))
		for j, ax := range axes {
			axisDTOs[j] = RadarAxis{Label: ax.Label, Value: ax.Value}
		}

		var ranking []RankingItemModel
		s.db.Where("radar_section_id = ?", rs.ID).Order("\"rank\" ASC").Find(&ranking)
		rankingDTOs := make([]RankingItem, len(ranking))
		for j, rk := range ranking {
			rankingDTOs[j] = RankingItem{Rank: rk.Rank, Label: rk.Label}
		}

		sectionDTOs[i] = RadarSection{
			Mode: rs.Mode, Axes: axisDTOs, Ranking: rankingDTOs,
		}
	}

	months := strings.Split(m.HeatmapMonths, "\n")
	rows := strings.Split(m.HeatmapRows, "\n")
	if len(months) == 1 && months[0] == "" {
		months = nil
	}
	if len(rows) == 1 && rows[0] == "" {
		rows = nil
	}

	var entries []HeatmapEntryModel
	s.db.Where("stats_id = ?", m.ID).Order("row_index ASC, col_index ASC").Find(&entries)

	var heatmapValues [][]float32
	if len(entries) > 0 {
		maxRow := 0
		maxCol := 0
		for _, e := range entries {
			if e.RowIndex > maxRow {
				maxRow = e.RowIndex
			}
			if e.ColIndex > maxCol {
				maxCol = e.ColIndex
			}
		}
		heatmapValues = make([][]float32, maxRow+1)
		for i := range heatmapValues {
			heatmapValues[i] = make([]float32, maxCol+1)
		}
		for _, e := range entries {
			heatmapValues[e.RowIndex][e.ColIndex] = e.Value
		}
	}

	return Stats{
		Title: m.Title, Subtitle: m.Subtitle,
		Metrics: metricDTOs, RadarSections: sectionDTOs,
		HeatmapMonths: months, HeatmapRows: rows,
		HeatmapValues: heatmapValues,
	}
}

func (s *SQLiteStore) AvatarPresets() []AvatarPreset {
	var models []AvatarPresetModel
	s.db.Find(&models)
	result := make([]AvatarPreset, len(models))
	for i, m := range models {
		result[i] = AvatarPreset{
			ID: m.ID, TopColorHex: m.TopColorHex,
			BottomColorHex: m.BottomColorHex, ImageURL: m.ImageURL,
		}
	}
	return result
}

func (s *SQLiteStore) OwnProfile() Profile {
	var m ProfileModel
	if err := s.db.Where("type = ?", "own").First(&m).Error; err != nil {
		return Profile{}
	}
	return s.profileModelToDTO(m)
}

func (s *SQLiteStore) SetOwnProfile(p Profile) {
	var m ProfileModel
	s.db.Where("type = ?", "own").First(&m)

	s.db.Model(&m).Updates(map[string]any{
		"name": p.Name, "handle": p.Handle,
		"quote": p.Quote, "completed_books": p.CompletedBooks,
	})

	s.db.Model(&AvatarModel{}).Where("profile_id = ?", m.ID).Updates(map[string]any{
		"top_color_hex":    p.Avatar.TopColorHex,
		"bottom_color_hex": p.Avatar.BottomColorHex,
		"avatar_preset_id": p.Avatar.AvatarPresetID,
		"preset_image_url": p.Avatar.PresetImageURL,
		"image_uri":        p.Avatar.ImageURI,
	})

	if p.ReadingGoal != nil {
		var goal ReadingGoalModel
		err := s.db.Where("profile_id = ?", m.ID).First(&goal).Error
		if err != nil {
			s.db.Create(&ReadingGoalModel{
				ProfileID:                 m.ID,
				TargetPagesPerDay:         p.ReadingGoal.TargetPagesPerDay,
				CurrentAveragePagesPerDay: p.ReadingGoal.CurrentAveragePagesPerDay,
			})
		} else {
			s.db.Model(&goal).Updates(map[string]any{
				"target_pages_per_day":          p.ReadingGoal.TargetPagesPerDay,
				"current_average_pages_per_day": p.ReadingGoal.CurrentAveragePagesPerDay,
			})
		}
	}
}

func (s *SQLiteStore) PublicProfile() Profile {
	var m ProfileModel
	if err := s.db.Where("type = ?", "public").First(&m).Error; err != nil {
		return Profile{}
	}
	return s.profileModelToDTO(m)
}

func (s *SQLiteStore) Stats() Stats {
	var profile ProfileModel
	if err := s.db.Where("type = ?", "own").First(&profile).Error; err != nil {
		return Stats{}
	}
	var m StatsModel
	if err := s.db.Where("profile_id = ?", profile.ID).First(&m).Error; err != nil {
		return Stats{}
	}
	return s.statsModelToDTO(m)
}

func (s *SQLiteStore) Notifications() Notifications {
	var models []NotificationEntryModel
	s.db.Find(&models)
	items := make([]NotificationEntry, len(models))
	for i, m := range models {
		n := NotificationEntry{
			ID: m.ID, Type: m.Type, Timestamp: m.Timestamp,
			Actor: NotificationActor{
				Name:                 m.ActorName,
				AvatarTopColorHex:    m.ActorTopHex,
				AvatarBottomColorHex: m.ActorBottomHex,
			},
			FallbackText: m.FallbackText,
		}
		if m.BookTitle != nil {
			n.Book = &NotificationBookSummary{
				Title: *m.BookTitle, Author: *m.BookAuthor,
				CoverImageURL: *m.BookCoverURL,
			}
		}
		items[i] = n
	}
	return Notifications{Title: "Activity and Notifications", Items: items}
}

func (s *SQLiteStore) Books() []Book {
	var models []BookModel
	s.db.Find(&models)
	result := make([]Book, len(models))
	for i, m := range models {
		result[i] = Book{
			ID: m.ID, Title: m.Title, Author: m.Author,
			Description: m.Description, CoverImageURL: m.CoverImageURL,
			Pages: m.Pages, ISBN: m.ISBN, Genre: m.Genre,
			Language: m.Language, Published: m.Published,
		}
	}
	return result
}

func (s *SQLiteStore) BookByID(bookID string) (Book, bool) {
	var m BookModel
	if err := s.db.First(&m, "id = ?", bookID).Error; err != nil {
		return Book{}, false
	}
	return Book{
		ID: m.ID, Title: m.Title, Author: m.Author,
		Description: m.Description, CoverImageURL: m.CoverImageURL,
		Pages: m.Pages, ISBN: m.ISBN, Genre: m.Genre,
		Language: m.Language, Published: m.Published,
	}, true
}

func (s *SQLiteStore) ExploreUsers() []ExploreUser {
	var models []ExploreUserModel
	s.db.Find(&models)
	result := make([]ExploreUser, len(models))
	for i, m := range models {
		result[i] = ExploreUser{
			ID: m.ID, Name: m.Name, Handle: m.Handle, Bio: m.Bio,
			AvatarImageURL:       m.AvatarImageURL,
			AvatarTopColorHex:    m.AvatarTopColorHex,
			AvatarBottomColorHex: m.AvatarBottomColorHex,
			FollowersLabel:       m.FollowersLabel,
			BooksReadLabel:       m.BooksReadLabel,
		}
	}
	return result
}

func (s *SQLiteStore) ReadingProgress(bookID string) (ReadingProgress, bool) {
	var m ReadingProgressModel
	if err := s.db.First(&m, "book_id = ?", bookID).Error; err != nil {
		return ReadingProgress{}, false
	}
	return ReadingProgress{
		BookID: m.BookID, CurrentPage: m.CurrentPage,
		TotalPages: m.TotalPages, UpdatedAt: m.UpdatedAt,
	}, true
}

func (s *SQLiteStore) SetReadingProgress(bookID string, currentPage int, totalPages int) error {
	var book BookModel
	if err := s.db.First(&book, "id = ?", bookID).Error; err != nil {
		return ErrNotFound
	}
	model := ReadingProgressModel{
		BookID:      bookID,
		CurrentPage: currentPage,
		TotalPages:  totalPages,
		UpdatedAt:   time.Now().Format("02/01/2006"),
	}
	s.db.Save(&model)
	return nil
}

func (s *SQLiteStore) Reviews(bookID string, page int, limit int) ([]Review, int) {
	var models []ReviewModel
	var total int64
	s.db.Model(&ReviewModel{}).Where("book_id = ?", bookID).Count(&total)
	s.db.Where("book_id = ?", bookID).
		Order("likes DESC, created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&models)
	result := make([]Review, len(models))
	for i, m := range models {
		var likedBy []string
		if m.LikedBy != "" {
			json.Unmarshal([]byte(m.LikedBy), &likedBy)
		}
		result[i] = Review{
			ID: m.ID, UserID: m.UserID, ReviewerName: m.ReviewerName,
			ReviewerAvatar: m.ReviewerAvatar, Rating: m.Rating,
			Text: m.Text, Likes: m.Likes, LikedBy: likedBy, CreatedAt: m.CreatedAt,
		}
	}
	return result, int(total)
}

func (s *SQLiteStore) SaveReview(bookID string, review Review) error {
	likedByBytes, _ := json.Marshal(review.LikedBy)
	var existing ReviewModel
	err := s.db.Where("book_id = ? AND user_id = ?", bookID, review.UserID).First(&existing).Error
	if err == nil {
		return s.db.Model(&existing).Updates(map[string]interface{}{
			"rating":          review.Rating,
			"text":            review.Text,
			"reviewer_name":   review.ReviewerName,
			"reviewer_avatar": review.ReviewerAvatar,
			"likes":           review.Likes,
			"liked_by":        string(likedByBytes),
			"created_at":      review.CreatedAt,
		}).Error
	}
	return s.db.Create(&ReviewModel{
		ID: review.ID, BookID: bookID, UserID: review.UserID,
		ReviewerName: review.ReviewerName, ReviewerAvatar: review.ReviewerAvatar,
		Rating: review.Rating, Text: review.Text,
		Likes: review.Likes, LikedBy: string(likedByBytes), CreatedAt: review.CreatedAt,
	}).Error
}

func (s *SQLiteStore) ToggleLike(userID string, reviewID string) (Review, error) {
	var existing ReviewLikeModel
	err := s.db.Where("review_id = ? AND user_id = ?", reviewID, userID).First(&existing).Error
	if err == nil {
		s.db.Delete(&existing)
	} else {
		s.db.Create(&ReviewLikeModel{ReviewID: reviewID, UserID: userID})
	}

	var count int64
	s.db.Model(&ReviewLikeModel{}).Where("review_id = ?", reviewID).Count(&count)

	var review ReviewModel
	if err := s.db.Where("id = ?", reviewID).First(&review).Error; err != nil {
		return Review{}, ErrNotFound
	}

	var userIDs []string
	s.db.Model(&ReviewLikeModel{}).Where("review_id = ?", reviewID).Pluck("user_id", &userIDs)
	likedByBytes, _ := json.Marshal(userIDs)

	s.db.Model(&review).Updates(map[string]interface{}{
		"likes":    int(count),
		"liked_by": string(likedByBytes),
	})

	review.Likes = int(count)
	review.LikedBy = string(likedByBytes)

	return Review{
		ID: review.ID, UserID: review.UserID,
		ReviewerName: review.ReviewerName, ReviewerAvatar: review.ReviewerAvatar,
		Rating: review.Rating, Text: review.Text,
		Likes: review.Likes, LikedBy: userIDs, CreatedAt: review.CreatedAt,
	}, nil
}

func (s *SQLiteStore) Authors() []Author {
	var models []AuthorModel
	s.db.Preload("Books").Find(&models)
	result := make([]Author, len(models))
	for i, m := range models {
		bookDTOs := make([]Book, len(m.Books))
		for j, b := range m.Books {
			bookDTOs[j] = Book{
				ID: b.ID, Title: b.Title, Author: b.Author,
				Description: b.Description, CoverImageURL: b.CoverImageURL,
				Pages: b.Pages, ISBN: b.ISBN, Genre: b.Genre,
				Language: b.Language, Published: b.Published,
			}
		}
		result[i] = Author{
			ID: m.ID, Name: m.Name, BirthYear: m.BirthYear,
			DeathYear: m.DeathYear, Nationality: m.Nationality,
			Description: m.Description, Biography: m.Biography,
			ImageURL: m.ImageURL, Books: bookDTOs, Followers: m.Followers,
		}
	}
	return result
}

func (s *SQLiteStore) AuthorByID(authorID string) (Author, bool) {
	var m AuthorModel
	if err := s.db.Preload("Books").First(&m, "id = ?", authorID).Error; err != nil {
		return Author{}, false
	}
	bookDTOs := make([]Book, len(m.Books))
	for j, b := range m.Books {
		bookDTOs[j] = Book{
			ID: b.ID, Title: b.Title, Author: b.Author,
			Description: b.Description, CoverImageURL: b.CoverImageURL,
			Pages: b.Pages, ISBN: b.ISBN, Genre: b.Genre,
			Language: b.Language, Published: b.Published,
		}
	}
	return Author{
		ID: m.ID, Name: m.Name, BirthYear: m.BirthYear,
		DeathYear: m.DeathYear, Nationality: m.Nationality,
		Description: m.Description, Biography: m.Biography,
		ImageURL: m.ImageURL, Books: bookDTOs, Followers: m.Followers,
	}, true
}

func (s *SQLiteStore) ToggleFollow(authorID string) bool {
	var profile ProfileModel
	s.db.Where("type = ?", "own").First(&profile)

	var existing FollowedAuthorModel
	err := s.db.Where("profile_id = ? AND author_id = ?", profile.ID, authorID).First(&existing).Error
	if err == nil {
		s.db.Delete(&existing)
		return false
	}
	s.db.Create(&FollowedAuthorModel{ProfileID: profile.ID, AuthorID: authorID})
	return true
}

func (s *SQLiteStore) AddBookToLibrary(bookID string) error {
	var book BookModel
	if err := s.db.First(&book, "id = ?", bookID).Error; err != nil {
		return ErrNotFound
	}

	var profile ProfileModel
	s.db.Where("type = ?", "own").First(&profile)

	var count int64
	s.db.Model(&PublicLibraryModel{}).
		Where("profile_id = ? AND title = ?", profile.ID, book.Title).
		Count(&count)
	if count > 0 {
		return ErrAlreadyInLibrary
	}

	s.db.Create(&PublicLibraryModel{
		ProfileID:     profile.ID,
		BookID:        book.ID,
		Title:         book.Title,
		CoverImageURL: book.CoverImageURL,
	})
	return nil
}

func (s *SQLiteStore) RemoveBookFromLibrary(bookID string) error {
	var book BookModel
	if err := s.db.First(&book, "id = ?", bookID).Error; err != nil {
		return ErrNotFound
	}

	var profile ProfileModel
	s.db.Where("type = ?", "own").First(&profile)

	result := s.db.Where("profile_id = ? AND book_id = ?", profile.ID, bookID).Delete(&PublicLibraryModel{})
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return result.Error
}

var _ Storer = (*SQLiteStore)(nil)
