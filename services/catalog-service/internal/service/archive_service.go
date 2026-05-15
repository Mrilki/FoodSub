package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Mrilki/catalog-service/internal/model"
	"github.com/Mrilki/catalog-service/internal/repository"
	"github.com/Mrilki/catalog-service/pkg/logger"

	"go.uber.org/zap"
)

type ArchiveService interface {
	ArchiveOldMenus(ctx context.Context, olderThanDays int) error
	GetArchive(ctx context.Context, filename string) ([]byte, error)
	ListArchives(ctx context.Context) ([]string, error)
}

type archiveService struct {
	menuRepo  repository.MenuRepository
	minioRepo repository.MinIORepository
	log       *logger.Logger
}

func NewArchiveService(
	menuRepo repository.MenuRepository,
	minioRepo repository.MinIORepository,
	log *logger.Logger,
) ArchiveService {
	return &archiveService{
		menuRepo:  menuRepo,
		minioRepo: minioRepo,
		log:       log,
	}
}

func (s *archiveService) ArchiveOldMenus(ctx context.Context, olderThanDays int) error {
	s.log.Info("Starting archive process", zap.Int("older_than_days", olderThanDays))

	allMenus, err := s.menuRepo.GetAll(ctx)
	if err != nil {
		s.log.Error("Failed to get menus for archiving", zap.Error(err))
		return err
	}

	cutoffDate := time.Now().AddDate(0, 0, -olderThanDays)
	var oldMenus []*model.MenuItem

	for _, menu := range allMenus {
		if menu.UpdatedAt.Before(cutoffDate) && !menu.IsAvailable {
			oldMenus = append(oldMenus, menu)
		}
	}

	if len(oldMenus) == 0 {
		s.log.Info("No menus to archive")
		return nil
	}

	s.log.Info("Found menus to archive", zap.Int("count", len(oldMenus)))

	csvData, err := repository.CreateCSVArchive(oldMenus)
	if err != nil {
		s.log.Error("Failed to create CSV archive", zap.Error(err))
		return err
	}

	jsonData, err := repository.CreateJSONArchive(oldMenus)
	if err != nil {
		s.log.Error("Failed to create JSON archive", zap.Error(err))
		return err
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	csvFilename := fmt.Sprintf("archives/menus_%s.csv", timestamp)
	jsonFilename := fmt.Sprintf("archives/menus_%s.json", timestamp)

	if err := s.minioRepo.UploadArchive(ctx, csvFilename, csvData); err != nil {
		return err
	}

	if err := s.minioRepo.UploadArchive(ctx, jsonFilename, jsonData); err != nil {
		return err
	}

	s.log.Info("Archive process completed",
		zap.Int("archived_count", len(oldMenus)),
		zap.String("csv_file", csvFilename),
		zap.String("json_file", jsonFilename))

	return nil
}

func (s *archiveService) GetArchive(ctx context.Context, filename string) ([]byte, error) {
	return s.minioRepo.GetArchive(ctx, filename)
}

func (s *archiveService) ListArchives(ctx context.Context) ([]string, error) {
	return s.minioRepo.ListArchives(ctx, "archives/")
}
