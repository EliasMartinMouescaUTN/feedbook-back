package feedbook

type Book struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Author        string `json:"author"`
	Description   string `json:"description"`
	CoverImageURL string `json:"cover_image_url,omitempty"`
	Pages         int    `json:"pages"`
	ISBN          string `json:"isbn"`
	Genre         string `json:"genre"`
	Language      string `json:"language"`
	Published     string `json:"published"`
}

type ReadingProgress struct {
	BookID      string `json:"book_id"`
	CurrentPage int    `json:"current_page"`
	TotalPages  int    `json:"total_pages"`
	UpdatedAt   string `json:"updated_at"`
}

type Review struct {
	ID             string   `json:"id"`
	UserID         string   `json:"user_id"`
	ReviewerName   string   `json:"reviewer_name"`
	ReviewerAvatar string   `json:"reviewer_avatar,omitempty"`
	Rating         float32  `json:"rating"`
	Text           string   `json:"text"`
	Likes          int      `json:"likes"`
	LikedBy        []string `json:"liked_by"`
	CreatedAt      string   `json:"created_at"`
}

type ExploreUser struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	Handle               string `json:"handle"`
	Bio                  string `json:"bio"`
	AvatarImageURL       string `json:"avatarImageUrl,omitempty"`
	AvatarTopColorHex    int64  `json:"avatarTopColorHex"`
	AvatarBottomColorHex int64  `json:"avatarBottomColorHex"`
	FollowersLabel       string `json:"followersLabel"`
	BooksReadLabel       string `json:"booksReadLabel"`
}

type Author struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	BirthYear   int    `json:"birth_year"`
	DeathYear   *int   `json:"death_year"`
	Nationality string `json:"nationality"`
	Description string `json:"description"`
	Biography   string `json:"biography"`
	ImageURL    string `json:"image_url,omitempty"`
	Books       []Book `json:"books"`
	Followers   int    `json:"followers"`
}

type Home struct {
	TrendingTitle string            `json:"trendingTitle"`
	Avatar        Avatar            `json:"avatar"`
	FeaturedBook  HomeFeaturedBook  `json:"featuredBook"`
	RankedBooks   []HomeRankedBook  `json:"rankedBooks"`
	ReadingRooms  []HomeReadingRoom `json:"readingRooms"`
	Curators      []HomeCurator     `json:"curators"`
}

type HomeFeaturedBook struct {
	Label         string `json:"label"`
	Title         string `json:"title"`
	Author        string `json:"author"`
	CoverImageURL string `json:"coverImageUrl,omitempty"`
}

type HomeRankedBook struct {
	RankLabel     string `json:"rankLabel"`
	Title         string `json:"title"`
	Author        string `json:"author"`
	CoverImageURL string `json:"coverImageUrl,omitempty"`
}

type HomeReadingRoom struct {
	HostName         string `json:"hostName"`
	HostImageURL     string `json:"hostImageUrl,omitempty"`
	Title            string `json:"title"`
	ReaderCountLabel string `json:"readerCountLabel"`
}

type HomeCurator struct {
	Name     string `json:"name"`
	Focus    string `json:"focus"`
	ImageURL string `json:"imageUrl,omitempty"`
}

type Library struct {
	Title          string        `json:"title"`
	Subtitle       string        `json:"subtitle"`
	Avatar         Avatar        `json:"avatar"`
	CurrentBook    CurrentBook   `json:"currentBook"`
	ReadingBooks   []LibraryBook `json:"readingBooks"`
	ShelfBooks     []LibraryBook `json:"shelfBooks"`
	CompletedBooks int           `json:"completedBooks"`
	ReadHistory    []ReadBook    `json:"readHistory"`
}

type ReadBook struct {
	Title          string `json:"title"`
	Author         string `json:"author"`
	StartedOn      string `json:"startedOn"`
	FinishedOn     string `json:"finishedOn"`
	PersonalRating int    `json:"personalRating"`
	CoverAccentHex int64  `json:"coverAccentHex"`
}

type Profile struct {
	Name                   string           `json:"name"`
	Handle                 string           `json:"handle"`
	Quote                  string           `json:"quote"`
	Avatar                 Avatar           `json:"avatar"`
	AvailableAvatarPresets []AvatarPreset   `json:"availableAvatarPresets"`
	ReadingGoal            *ReadingGoal     `json:"readingGoal"`
	ReadingStreak          ReadingStreak    `json:"readingStreak"`
	CurrentBook            CurrentBook      `json:"currentBook"`
	UpNextBooks            []QueuedBook     `json:"upNextBooks"`
	CompletedBooks         int              `json:"completedBooks"`
	ProfileStats           []ProfileStat    `json:"profileStats"`
	PublicLibrary          []LibraryBook    `json:"publicLibrary"`
	FeaturedReviews        []FeaturedReview `json:"featuredReviews"`
}

type Avatar struct {
	TopColorHex    int64  `json:"topColorHex"`
	BottomColorHex int64  `json:"bottomColorHex"`
	AvatarPresetID string `json:"avatarPresetId,omitempty"`
	PresetImageURL string `json:"presetImageUrl,omitempty"`
	ImageURI       string `json:"imageUri,omitempty"`
}

type AvatarPreset struct {
	ID             string `json:"id"`
	TopColorHex    int64  `json:"topColorHex"`
	BottomColorHex int64  `json:"bottomColorHex"`
	ImageURL       string `json:"imageUrl,omitempty"`
}

type ReadingGoal struct {
	TargetPagesPerDay         int `json:"targetPagesPerDay"`
	CurrentAveragePagesPerDay int `json:"currentAveragePagesPerDay"`
}

type ReadingStreak struct {
	Days int         `json:"days"`
	Week []StreakDay `json:"week"`
}

type StreakDay struct {
	Label        string  `json:"label"`
	FillFraction float32 `json:"fillFraction"`
	IsToday      bool    `json:"isToday"`
	Completed    bool    `json:"completed"`
}

type CurrentBook struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Author        string  `json:"author"`
	Page          int     `json:"page"`
	TotalPages    int     `json:"totalPages"`
	Progress      float32 `json:"progress"`
	CoverImageURL string  `json:"coverImageUrl,omitempty"`
}

type QueuedBook struct {
	Title         string `json:"title"`
	Author        string `json:"author"`
	CoverImageURL string `json:"coverImageUrl,omitempty"`
}

type ProfileStat struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type LibraryBook struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	CoverImageURL string `json:"coverImageUrl,omitempty"`
}

type FeaturedReview struct {
	BookTitle     string `json:"bookTitle"`
	Rating        int    `json:"rating"`
	TimeAgo       string `json:"timeAgo"`
	Excerpt       string `json:"excerpt"`
	CoverImageURL string `json:"coverImageUrl,omitempty"`
}

type UpdateProfileRequest struct {
	Name                 string  `json:"name"`
	Handle               string  `json:"handle"`
	Quote                string  `json:"quote"`
	AvatarTopColorHex    int64   `json:"avatarTopColorHex"`
	AvatarBottomColorHex int64   `json:"avatarBottomColorHex"`
	AvatarPresetID       *string `json:"avatarPresetId"`
	AvatarImageURI       *string `json:"avatarImageUri"`
	TargetPagesPerDay    *int    `json:"targetPagesPerDay"`
}

type Stats struct {
	Title         string         `json:"title"`
	Subtitle      string         `json:"subtitle"`
	Metrics       []StatsMetric  `json:"metrics"`
	HeatmapMonths []string       `json:"heatmapMonths"`
	HeatmapRows   []string       `json:"heatmapRows"`
	HeatmapValues [][]float32    `json:"heatmapValues"`
	RadarSections []RadarSection `json:"radarSections"`
}

type StatsMetric struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type RadarSection struct {
	Mode    string        `json:"mode"`
	Axes    []RadarAxis   `json:"axes"`
	Ranking []RankingItem `json:"ranking"`
}

type RadarAxis struct {
	Label string  `json:"label"`
	Value float32 `json:"value"`
}

type RankingItem struct {
	Rank  int    `json:"rank"`
	Label string `json:"label"`
}

type Notifications struct {
	Title string              `json:"title"`
	Items []NotificationEntry `json:"items"`
}

type NotificationEntry struct {
	ID           string                   `json:"id"`
	Type         string                   `json:"type"`
	Timestamp    string                   `json:"timestamp"`
	Actor        NotificationActor        `json:"actor"`
	Book         *NotificationBookSummary `json:"book,omitempty"`
	FallbackText string                   `json:"fallbackText"`
}

type NotificationActor struct {
	Name                 string `json:"name"`
	AvatarTopColorHex    int64  `json:"avatarTopColorHex"`
	AvatarBottomColorHex int64  `json:"avatarBottomColorHex"`
}

type NotificationBookSummary struct {
	Title         string `json:"title"`
	Author        string `json:"author"`
	CoverImageURL string `json:"coverImageUrl,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type RegisterPushTokenRequest struct {
	Token    string `json:"token"`
	Platform string `json:"platform"`
}

type SendPushRequest struct {
	Title string            `json:"title"`
	Body  string            `json:"body"`
	Data  map[string]string `json:"data,omitempty"`
	Token string            `json:"token,omitempty"`
}

type SendPushResponse struct {
	Sent    int      `json:"sent"`
	Failed  int      `json:"failed"`
	Message string   `json:"message,omitempty"`
	IDs     []string `json:"ids,omitempty"`
	Errors  []string `json:"errors,omitempty"`
}

type PushTokenInfo struct {
	Token    string `json:"token"`
	Platform string `json:"platform"`
}
