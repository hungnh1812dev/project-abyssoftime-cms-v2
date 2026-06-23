package locale

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

var codePattern = regexp.MustCompile(`^[a-z]+(-[a-z]+)*$`)

var knownNames = map[string]string{
	"en":    "English",
	"vi":    "Tiếng Việt",
	"ja":    "日本語",
	"ko":    "한국어",
	"zh":    "中文",
	"zh-cn": "简体中文",
	"zh-tw": "繁體中文",
	"fr":    "Français",
	"de":    "Deutsch",
	"es":    "Español",
	"pt":    "Português",
	"it":    "Italiano",
	"ru":    "Русский",
	"ar":    "العربية",
	"th":    "ไทย",
}

type CreateInput struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	IsDefault bool   `json:"isDefault"`
}

type UpdateInput struct {
	Name      *string `json:"name"`
	IsDefault *bool   `json:"isDefault"`
}

type UseCase struct {
	localeRepo repository.LocaleRepository
	docRepo    repository.DocumentRepository
	ctRepo     repository.ContentTypeRepository
}

func New(localeRepo repository.LocaleRepository, docRepo repository.DocumentRepository, ctRepo repository.ContentTypeRepository) *UseCase {
	return &UseCase{
		localeRepo: localeRepo,
		docRepo:    docRepo,
		ctRepo:     ctRepo,
	}
}

func (uc *UseCase) List(ctx context.Context) ([]*entity.Locale, error) {
	return uc.localeRepo.FindAll(ctx)
}

func (uc *UseCase) Create(ctx context.Context, input CreateInput) (*entity.Locale, error) {
	if err := validateCode(input.Code); err != nil {
		return nil, err
	}
	if err := validateName(input.Name); err != nil {
		return nil, err
	}

	_, err := uc.localeRepo.FindByCode(ctx, input.Code)
	if err == nil {
		return nil, fmt.Errorf("%w: locale code %q already exists", pkgerrors.ErrConflict, input.Code)
	}
	if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		return nil, err
	}

	if input.IsDefault {
		if err := uc.localeRepo.ClearDefault(ctx); err != nil {
			return nil, err
		}
	}

	now := time.Now()
	locale := &entity.Locale{
		Code:      input.Code,
		Name:      input.Name,
		IsDefault: input.IsDefault,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := uc.localeRepo.Create(ctx, locale); err != nil {
		return nil, err
	}
	return locale, nil
}

func (uc *UseCase) Update(ctx context.Context, code string, input UpdateInput) (*entity.Locale, error) {
	locale, err := uc.localeRepo.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		if err := validateName(*input.Name); err != nil {
			return nil, err
		}
		locale.Name = *input.Name
	}

	if input.IsDefault != nil {
		if *input.IsDefault {
			if err := uc.localeRepo.ClearDefault(ctx); err != nil {
				return nil, err
			}
			locale.IsDefault = true
		} else if locale.IsDefault {
			allLocales, err := uc.localeRepo.FindAll(ctx)
			if err != nil {
				return nil, err
			}
			defaultCount := 0
			for _, loc := range allLocales {
				if loc.IsDefault {
					defaultCount++
				}
			}
			if defaultCount <= 1 {
				return nil, fmt.Errorf("%w: cannot unset the only default locale", pkgerrors.ErrConflict)
			}
			locale.IsDefault = false
		}
	}

	locale.UpdatedAt = time.Now()
	if err := uc.localeRepo.Update(ctx, locale); err != nil {
		return nil, err
	}
	return locale, nil
}

func (uc *UseCase) Delete(ctx context.Context, code string) error {
	locale, err := uc.localeRepo.FindByCode(ctx, code)
	if err != nil {
		return err
	}

	allLocales, err := uc.localeRepo.FindAll(ctx)
	if err != nil {
		return err
	}
	if len(allLocales) <= 1 {
		return fmt.Errorf("%w: cannot delete the last locale", pkgerrors.ErrConflict)
	}

	docCount, err := uc.countDocumentsForLocale(ctx, code)
	if err != nil {
		return err
	}
	if docCount > 0 {
		return fmt.Errorf("%w: cannot delete locale %q — %d document(s) reference it", pkgerrors.ErrConflict, code, docCount)
	}

	if locale.IsDefault {
		for _, loc := range allLocales {
			if loc.Code != code {
				loc.IsDefault = true
				loc.UpdatedAt = time.Now()
				if err := uc.localeRepo.Update(ctx, loc); err != nil {
					return err
				}
				break
			}
		}
	}

	return uc.localeRepo.Delete(ctx, code)
}

func (uc *UseCase) Seed(ctx context.Context, envLocales []string) error {
	existing, err := uc.localeRepo.FindAll(ctx)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return nil
	}

	codes := envLocales
	if len(codes) == 0 {
		codes = []string{"en"}
	}

	now := time.Now()
	for idx, code := range codes {
		name := code
		if known, ok := knownNames[code]; ok {
			name = known
		}
		locale := &entity.Locale{
			Code:      code,
			Name:      name,
			IsDefault: idx == 0,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := uc.localeRepo.Create(ctx, locale); err != nil {
			return fmt.Errorf("seed locale %q: %w", code, err)
		}
	}
	return nil
}

func (uc *UseCase) countDocumentsForLocale(ctx context.Context, locale string) (int64, error) {
	contentTypes, err := uc.ctRepo.FindAll(ctx)
	if err != nil {
		return 0, err
	}
	var total int64
	for _, contentType := range contentTypes {
		count, err := uc.docRepo.CountByLocale(ctx, contentType.Slug, locale)
		if err != nil {
			return 0, err
		}
		total += count
	}
	return total, nil
}

func validateCode(code string) error {
	if len(code) < 2 || len(code) > 5 {
		return fmt.Errorf("%w: locale code must be 2-5 characters", pkgerrors.ErrValidation)
	}
	if !codePattern.MatchString(code) {
		return fmt.Errorf("%w: locale code must be lowercase letters and hyphens", pkgerrors.ErrValidation)
	}
	return nil
}

func validateName(name string) error {
	if len(name) == 0 || len(name) > 100 {
		return fmt.Errorf("%w: locale name must be 1-100 characters", pkgerrors.ErrValidation)
	}
	return nil
}
