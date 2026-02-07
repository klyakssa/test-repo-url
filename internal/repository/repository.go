package repository

type Repository interface {
	Load() map[string]string
	Save(map[string]string) error
	Shorten(string) (string, error)
	Unshorten(string) (string, error)
}

type FileStorageRepository interface {
	Save(map[string]string) error
	Load() (map[string]string, error)
	Delete() error
	Close() error
}
