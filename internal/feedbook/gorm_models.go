package feedbook

import "gorm.io/gorm"

type AccountModel struct {
	Username string `gorm:"primaryKey"`
	Password string
}

type ProfileModel struct {
	gorm.Model
	Type            string `gorm:"uniqueIndex:idx_profile_type;size:10"`
	Name            string
	Handle          string
	Quote           string
	CompletedBooks  int
	Avatar          AvatarModel           `gorm:"foreignKey:ProfileID;constraint:OnDelete:CASCADE"`
	ReadingGoal     *ReadingGoalModel     `gorm:"foreignKey:ProfileID;constraint:OnDelete:CASCADE"`
	ReadingStreak   *ReadingStreakModel   `gorm:"foreignKey:ProfileID;constraint:OnDelete:CASCADE"`
	CurrentBook     *CurrentBookModel     `gorm:"foreignKey:ProfileID;constraint:OnDelete:CASCADE"`
	QueuedBooks     []QueuedBookModel     `gorm:"foreignKey:ProfileID;constraint:OnDelete:CASCADE"`
	ProfileStats    []ProfileStatModel    `gorm:"foreignKey:ProfileID;constraint:OnDelete:CASCADE"`
	PublicLibrary   []PublicLibraryModel  `gorm:"foreignKey:ProfileID;constraint:OnDelete:CASCADE"`
	FeaturedReviews []FeaturedReviewModel `gorm:"foreignKey:ProfileID;constraint:OnDelete:CASCADE"`
}

type AvatarModel struct {
	gorm.Model
	ProfileID      uint
	TopColorHex    int64
	BottomColorHex int64
	AvatarPresetID string
	PresetImageURL string
	ImageURI       string
}

type ReadingGoalModel struct {
	gorm.Model
	ProfileID                 uint `gorm:"uniqueIndex"`
	TargetPagesPerDay         int
	CurrentAveragePagesPerDay int
}

type ReadingStreakModel struct {
	gorm.Model
	ProfileID uint `gorm:"uniqueIndex"`
	DaysCount int
	Days      []StreakDayModel `gorm:"foreignKey:ReadingStreakID;constraint:OnDelete:CASCADE"`
}

type StreakDayModel struct {
	gorm.Model
	ReadingStreakID uint
	Label           string
	FillFraction    float32
	IsToday         bool
	Completed       bool
}

type CurrentBookModel struct {
	gorm.Model
	ProfileID     uint `gorm:"uniqueIndex"`
	BookID        string
	Title         string
	Author        string
	Page          int
	TotalPages    int
	Progress      float32
	CoverImageURL string
}

type QueuedBookModel struct {
	gorm.Model
	ProfileID     uint
	Title         string
	Author        string
	CoverImageURL string
}

type ProfileStatModel struct {
	gorm.Model
	ProfileID uint
	Label     string
	Value     string
}

type PublicLibraryModel struct {
	gorm.Model
	ProfileID     uint
	BookID        string
	Title         string
	CoverImageURL string
}

type FeaturedReviewModel struct {
	gorm.Model
	ProfileID     uint
	BookTitle     string
	Rating        int
	TimeAgo       string
	Excerpt       string
	CoverImageURL string
}

type StatsModel struct {
	gorm.Model
	ProfileID      uint `gorm:"uniqueIndex"`
	Title          string
	Subtitle       string
	HeatmapMonths  string              `gorm:"type:text"`
	HeatmapRows    string              `gorm:"type:text"`
	Metrics        []StatsMetricModel  `gorm:"foreignKey:StatsID;constraint:OnDelete:CASCADE"`
	RadarSections  []RadarSectionModel `gorm:"foreignKey:StatsID;constraint:OnDelete:CASCADE"`
	HeatmapEntries []HeatmapEntryModel `gorm:"foreignKey:StatsID;constraint:OnDelete:CASCADE"`
}

type StatsMetricModel struct {
	gorm.Model
	StatsID uint
	Label   string
	Value   string
}

type RadarSectionModel struct {
	gorm.Model
	StatsID uint
	Mode    string
	Axes    []RadarAxisModel   `gorm:"foreignKey:RadarSectionID;constraint:OnDelete:CASCADE"`
	Ranking []RankingItemModel `gorm:"foreignKey:RadarSectionID;constraint:OnDelete:CASCADE"`
}

type RadarAxisModel struct {
	gorm.Model
	RadarSectionID uint
	Label          string
	Value          float32
}

type RankingItemModel struct {
	gorm.Model
	RadarSectionID uint
	Rank           int
	Label          string
}

type HeatmapEntryModel struct {
	gorm.Model
	StatsID  uint
	RowIndex int
	ColIndex int
	Value    float32
}

type BookModel struct {
	ID            string `gorm:"primaryKey"`
	Title         string
	Author        string
	Description   string
	CoverImageURL string
	Pages         int
	ISBN          string
	Genre         string
	Language      string
	Published     string
}

type ReviewModel struct {
	ID             string `gorm:"primaryKey"`
	BookID         string `gorm:"index"`
	UserID         string `gorm:"index"`
	ReviewerName   string
	ReviewerAvatar string
	Rating         float32
	Text           string
	Likes          int
	LikedBy        string
	CreatedAt      string
}

type ReviewLikeModel struct {
	ReviewID string `gorm:"primaryKey"`
	UserID   string `gorm:"primaryKey"`
}

type ReadingProgressModel struct {
	BookID      string `gorm:"primaryKey"`
	CurrentPage int
	TotalPages  int
	UpdatedAt   string
}

type AuthorModel struct {
	ID          string `gorm:"primaryKey"`
	Name        string
	BirthYear   int
	DeathYear   *int
	Nationality string
	Description string
	Biography   string
	ImageURL    string
	Followers   int
	Books       []BookModel `gorm:"many2many:author_books;"`
}

type ExploreUserModel struct {
	ID                   string `gorm:"primaryKey"`
	Name                 string
	Handle               string
	Bio                  string
	AvatarImageURL       string
	AvatarTopColorHex    int64
	AvatarBottomColorHex int64
	FollowersLabel       string
	BooksReadLabel       string
}

type AvatarPresetModel struct {
	ID             string `gorm:"primaryKey"`
	TopColorHex    int64
	BottomColorHex int64
	ImageURL       string
}

type NotificationEntryModel struct {
	ID             string `gorm:"primaryKey"`
	Type           string
	Timestamp      string
	ActorName      string
	ActorTopHex    int64
	ActorBottomHex int64
	BookTitle      *string
	BookAuthor     *string
	BookCoverURL   *string
	FallbackText   string
}

type FollowedAuthorModel struct {
	ProfileID uint   `gorm:"primaryKey"`
	AuthorID  string `gorm:"primaryKey"`
}
