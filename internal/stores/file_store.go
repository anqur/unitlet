package stores

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/anqur/unitlet/pkg/errs"
	"github.com/anqur/unitlet/pkg/units"
)

const DefaultFileStorePath = "/opt/unitlet/units"

type FileStore struct {
	path string
}

func NewFileStore(path string) (units.Store, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, err
	}
	return &FileStore{path}, nil
}

func (s *FileStore) Location(name units.Name) units.Location {
	return units.Location(filepath.Join(DefaultFileStorePath, string(name)))
}

func (s *FileStore) GetUnit(_ context.Context, name units.Name) (*units.Unit, error) {
	data, err := os.ReadFile(s.filepath(name))
	if err != nil {
		return nil, err
	}
	ret := new(units.Unit)
	return ret, ret.Unmarshal(data)
}

func (s *FileStore) CreateUnits(ctx context.Context, us []*units.Unit) error {
	return s.writeUnits(ctx, us, false)
}

func (s *FileStore) DeleteUnit(_ context.Context, name units.Name) error {
	return os.Remove(s.filepath(name))
}

func (s *FileStore) UpdateUnits(ctx context.Context, us []*units.Unit) error {
	return s.writeUnits(ctx, us, true)
}

func (s *FileStore) filepath(name units.Name) string {
	return filepath.Join(s.path, string(name))
}

func (s *FileStore) writeUnits(_ context.Context, us []*units.Unit, overwrite bool) error {
	for _, u := range us {
		path := s.filepath(u.ID.Name())

		if !overwrite {
			if _, err := os.Stat(path); err == nil {
				return fmt.Errorf("%w: %s", errs.ErrUnitFileExists, path)
			}
		}

		data, err := u.Marshal()
		if err != nil {
			return fmt.Errorf("%w: %v", errs.ErrMarshalUnitFile, err)
		}

		if err := os.WriteFile(path, data, 0644); err != nil {
			return fmt.Errorf("%w: %v", errs.ErrWriteUnitFile, err)
		}
	}
	return nil
}
