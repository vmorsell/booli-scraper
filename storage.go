package main

// Storage defines the interface for a storage adapter.
type Storage interface {
	Put(apt Apartment) error
}

func ensureDir(dir string, perm fs.FileMode) error {
	if _, err := os.Stat(dir); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("stat: %w", err)
		}

		if err := os.Mkdir(dir, perm); err != nil {
			return fmt.Errorf("mkdir: %w", err)
		}
	}
	return nil
}
