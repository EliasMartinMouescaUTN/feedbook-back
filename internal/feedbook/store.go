package feedbook

type Storer interface {
	AvatarPresets() []AvatarPreset
	OwnProfile() Profile
	SetOwnProfile(Profile)
	PublicProfile() Profile
	Stats() Stats
	Notifications() Notifications
	Books() []Book
	BookByID(string) (Book, bool)
	ExploreUsers() []ExploreUser
	ReadingProgress(string) (ReadingProgress, bool)
	SetReadingProgress(bookID string, currentPage int, totalPages int) error
	Reviews(bookID string, page int, limit int) ([]Review, int)
	SaveReview(bookID string, review Review) error
	ToggleLike(userID string, reviewID string) (Review, error)
	Authors() []Author
	AuthorByID(string) (Author, bool)
	IsFollowing(userID string, authorID string) bool
	ToggleFollow(string) bool
	AddBookToLibrary(bookID string) error
	RemoveBookFromLibrary(bookID string) error
}
