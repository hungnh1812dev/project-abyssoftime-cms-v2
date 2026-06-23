package locale

import (
	"context"
	"testing"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository/mock"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

func newTestUC() (*UseCase, *mock.LocaleRepository, *mock.DocumentRepository, *mock.ContentTypeRepository) {
	localeRepo := &mock.LocaleRepository{}
	docRepo := &mock.DocumentRepository{}
	ctRepo := &mock.ContentTypeRepository{}
	useCase := New(localeRepo, docRepo, ctRepo)
	return useCase, localeRepo, docRepo, ctRepo
}

func TestCreate_Success(t *testing.T) {
	useCase, localeRepo, _, _ := newTestUC()
	ctx := context.Background()

	localeRepo.FindByCodeFn = func(_ context.Context, _ string) (*entity.Locale, error) {
		return nil, pkgerrors.ErrNotFound
	}
	localeRepo.CreateFn = func(_ context.Context, _ *entity.Locale) error {
		return nil
	}

	locale, err := useCase.Create(ctx, CreateInput{Code: "en", Name: "English"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if locale.Code != "en" {
		t.Errorf("Code = %q, want %q", locale.Code, "en")
	}
	if locale.Name != "English" {
		t.Errorf("Name = %q, want %q", locale.Name, "English")
	}
}

func TestCreate_WithDefault_ClearsPrevious(t *testing.T) {
	useCase, localeRepo, _, _ := newTestUC()
	ctx := context.Background()

	cleared := false
	localeRepo.FindByCodeFn = func(_ context.Context, _ string) (*entity.Locale, error) {
		return nil, pkgerrors.ErrNotFound
	}
	localeRepo.ClearDefaultFn = func(_ context.Context) error {
		cleared = true
		return nil
	}
	localeRepo.CreateFn = func(_ context.Context, _ *entity.Locale) error {
		return nil
	}

	_, err := useCase.Create(ctx, CreateInput{Code: "vi", Name: "Tiếng Việt", IsDefault: true})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if !cleared {
		t.Error("expected ClearDefault to be called")
	}
}

func TestCreate_DuplicateCode_Conflict(t *testing.T) {
	useCase, localeRepo, _, _ := newTestUC()
	ctx := context.Background()

	localeRepo.FindByCodeFn = func(_ context.Context, _ string) (*entity.Locale, error) {
		return &entity.Locale{Code: "en"}, nil
	}

	_, err := useCase.Create(ctx, CreateInput{Code: "en", Name: "English"})
	if !pkgerrors.Is(err, pkgerrors.ErrConflict) {
		t.Errorf("err = %v, want ErrConflict", err)
	}
}

func TestCreate_InvalidCode_Short(t *testing.T) {
	useCase, _, _, _ := newTestUC()
	_, err := useCase.Create(context.Background(), CreateInput{Code: "e", Name: "English"})
	if !pkgerrors.Is(err, pkgerrors.ErrValidation) {
		t.Errorf("err = %v, want ErrValidation", err)
	}
}

func TestCreate_InvalidCode_TooLong(t *testing.T) {
	useCase, _, _, _ := newTestUC()
	_, err := useCase.Create(context.Background(), CreateInput{Code: "abcdef", Name: "Test"})
	if !pkgerrors.Is(err, pkgerrors.ErrValidation) {
		t.Errorf("err = %v, want ErrValidation", err)
	}
}

func TestCreate_InvalidCode_Uppercase(t *testing.T) {
	useCase, _, _, _ := newTestUC()
	_, err := useCase.Create(context.Background(), CreateInput{Code: "EN", Name: "English"})
	if !pkgerrors.Is(err, pkgerrors.ErrValidation) {
		t.Errorf("err = %v, want ErrValidation", err)
	}
}

func TestCreate_InvalidCode_LeadingHyphen(t *testing.T) {
	useCase, _, _, _ := newTestUC()
	_, err := useCase.Create(context.Background(), CreateInput{Code: "-en", Name: "English"})
	if !pkgerrors.Is(err, pkgerrors.ErrValidation) {
		t.Errorf("err = %v, want ErrValidation", err)
	}
}

func TestCreate_ValidCode_WithHyphen(t *testing.T) {
	useCase, localeRepo, _, _ := newTestUC()
	ctx := context.Background()

	localeRepo.FindByCodeFn = func(_ context.Context, _ string) (*entity.Locale, error) {
		return nil, pkgerrors.ErrNotFound
	}
	localeRepo.CreateFn = func(_ context.Context, _ *entity.Locale) error {
		return nil
	}

	locale, err := useCase.Create(ctx, CreateInput{Code: "zh-cn", Name: "简体中文"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if locale.Code != "zh-cn" {
		t.Errorf("Code = %q, want %q", locale.Code, "zh-cn")
	}
}

func TestCreate_EmptyName_Validation(t *testing.T) {
	useCase, _, _, _ := newTestUC()
	_, err := useCase.Create(context.Background(), CreateInput{Code: "en", Name: ""})
	if !pkgerrors.Is(err, pkgerrors.ErrValidation) {
		t.Errorf("err = %v, want ErrValidation", err)
	}
}

func TestUpdate_Name(t *testing.T) {
	useCase, localeRepo, _, _ := newTestUC()
	ctx := context.Background()

	localeRepo.FindByCodeFn = func(_ context.Context, _ string) (*entity.Locale, error) {
		return &entity.Locale{Code: "en", Name: "English", IsDefault: false}, nil
	}
	localeRepo.UpdateFn = func(_ context.Context, _ *entity.Locale) error {
		return nil
	}

	newName := "English (US)"
	updated, err := useCase.Update(ctx, "en", UpdateInput{Name: &newName})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != "English (US)" {
		t.Errorf("Name = %q, want %q", updated.Name, "English (US)")
	}
}

func TestUpdate_SetDefault(t *testing.T) {
	useCase, localeRepo, _, _ := newTestUC()
	ctx := context.Background()

	cleared := false
	localeRepo.FindByCodeFn = func(_ context.Context, _ string) (*entity.Locale, error) {
		return &entity.Locale{Code: "vi", Name: "Tiếng Việt", IsDefault: false}, nil
	}
	localeRepo.ClearDefaultFn = func(_ context.Context) error {
		cleared = true
		return nil
	}
	localeRepo.UpdateFn = func(_ context.Context, _ *entity.Locale) error {
		return nil
	}

	isDefault := true
	_, err := useCase.Update(ctx, "vi", UpdateInput{IsDefault: &isDefault})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if !cleared {
		t.Error("expected ClearDefault to be called")
	}
}

func TestUpdate_UnsetOnlyDefault_Conflict(t *testing.T) {
	useCase, localeRepo, _, _ := newTestUC()
	ctx := context.Background()

	localeRepo.FindByCodeFn = func(_ context.Context, _ string) (*entity.Locale, error) {
		return &entity.Locale{Code: "en", Name: "English", IsDefault: true}, nil
	}
	localeRepo.FindAllFn = func(_ context.Context) ([]*entity.Locale, error) {
		return []*entity.Locale{
			{Code: "en", Name: "English", IsDefault: true},
			{Code: "vi", Name: "Tiếng Việt", IsDefault: false},
		}, nil
	}

	isDefault := false
	_, err := useCase.Update(ctx, "en", UpdateInput{IsDefault: &isDefault})
	if !pkgerrors.Is(err, pkgerrors.ErrConflict) {
		t.Errorf("err = %v, want ErrConflict", err)
	}
}

func TestDelete_Success(t *testing.T) {
	useCase, localeRepo, docRepo, ctRepo := newTestUC()
	ctx := context.Background()

	localeRepo.FindByCodeFn = func(_ context.Context, _ string) (*entity.Locale, error) {
		return &entity.Locale{Code: "vi", Name: "Tiếng Việt", IsDefault: false}, nil
	}
	localeRepo.FindAllFn = func(_ context.Context) ([]*entity.Locale, error) {
		return []*entity.Locale{
			{Code: "en", IsDefault: true},
			{Code: "vi", IsDefault: false},
		}, nil
	}
	ctRepo.FindAllFn = func(_ context.Context) ([]*entity.ContentType, error) {
		return []*entity.ContentType{{Slug: "blog"}}, nil
	}
	docRepo.CountByLocaleFn = func(_ context.Context, _, _ string) (int64, error) {
		return 0, nil
	}
	localeRepo.DeleteFn = func(_ context.Context, _ string) error {
		return nil
	}

	if err := useCase.Delete(ctx, "vi"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestDelete_LastLocale_Conflict(t *testing.T) {
	useCase, localeRepo, _, _ := newTestUC()
	ctx := context.Background()

	localeRepo.FindByCodeFn = func(_ context.Context, _ string) (*entity.Locale, error) {
		return &entity.Locale{Code: "en", IsDefault: true}, nil
	}
	localeRepo.FindAllFn = func(_ context.Context) ([]*entity.Locale, error) {
		return []*entity.Locale{{Code: "en", IsDefault: true}}, nil
	}

	err := useCase.Delete(ctx, "en")
	if !pkgerrors.Is(err, pkgerrors.ErrConflict) {
		t.Errorf("err = %v, want ErrConflict", err)
	}
}

func TestDelete_HasDocuments_Conflict(t *testing.T) {
	useCase, localeRepo, docRepo, ctRepo := newTestUC()
	ctx := context.Background()

	localeRepo.FindByCodeFn = func(_ context.Context, _ string) (*entity.Locale, error) {
		return &entity.Locale{Code: "vi", IsDefault: false}, nil
	}
	localeRepo.FindAllFn = func(_ context.Context) ([]*entity.Locale, error) {
		return []*entity.Locale{{Code: "en"}, {Code: "vi"}}, nil
	}
	ctRepo.FindAllFn = func(_ context.Context) ([]*entity.ContentType, error) {
		return []*entity.ContentType{{Slug: "blog"}, {Slug: "page"}}, nil
	}
	docRepo.CountByLocaleFn = func(_ context.Context, slug, _ string) (int64, error) {
		if slug == "blog" {
			return 3, nil
		}
		return 2, nil
	}

	err := useCase.Delete(ctx, "vi")
	if !pkgerrors.Is(err, pkgerrors.ErrConflict) {
		t.Errorf("err = %v, want ErrConflict", err)
	}
}

func TestDelete_Default_ReassignsToNext(t *testing.T) {
	useCase, localeRepo, docRepo, ctRepo := newTestUC()
	ctx := context.Background()

	var reassignedCode string
	localeRepo.FindByCodeFn = func(_ context.Context, _ string) (*entity.Locale, error) {
		return &entity.Locale{Code: "en", IsDefault: true}, nil
	}
	localeRepo.FindAllFn = func(_ context.Context) ([]*entity.Locale, error) {
		return []*entity.Locale{
			{Code: "en", IsDefault: true},
			{Code: "vi", IsDefault: false},
		}, nil
	}
	ctRepo.FindAllFn = func(_ context.Context) ([]*entity.ContentType, error) {
		return nil, nil
	}
	docRepo.CountByLocaleFn = func(_ context.Context, _, _ string) (int64, error) {
		return 0, nil
	}
	localeRepo.UpdateFn = func(_ context.Context, loc *entity.Locale) error {
		if loc.IsDefault {
			reassignedCode = loc.Code
		}
		return nil
	}
	localeRepo.DeleteFn = func(_ context.Context, _ string) error {
		return nil
	}

	if err := useCase.Delete(ctx, "en"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if reassignedCode != "vi" {
		t.Errorf("reassigned default to %q, want %q", reassignedCode, "vi")
	}
}

func TestSeed_EmptyTable_WithEnvLocales(t *testing.T) {
	useCase, localeRepo, _, _ := newTestUC()
	ctx := context.Background()

	localeRepo.FindAllFn = func(_ context.Context) ([]*entity.Locale, error) {
		return nil, nil
	}

	var created []string
	var defaultCode string
	localeRepo.CreateFn = func(_ context.Context, loc *entity.Locale) error {
		created = append(created, loc.Code)
		if loc.IsDefault {
			defaultCode = loc.Code
		}
		return nil
	}

	if err := useCase.Seed(ctx, []string{"en", "vi"}); err != nil {
		t.Fatalf("Seed: %v", err)
	}
	if len(created) != 2 {
		t.Errorf("created %d locales, want 2", len(created))
	}
	if defaultCode != "en" {
		t.Errorf("default = %q, want %q", defaultCode, "en")
	}
}

func TestSeed_EmptyTable_NoEnvLocales(t *testing.T) {
	useCase, localeRepo, _, _ := newTestUC()
	ctx := context.Background()

	localeRepo.FindAllFn = func(_ context.Context) ([]*entity.Locale, error) {
		return nil, nil
	}

	var createdCode string
	localeRepo.CreateFn = func(_ context.Context, loc *entity.Locale) error {
		createdCode = loc.Code
		return nil
	}

	if err := useCase.Seed(ctx, nil); err != nil {
		t.Fatalf("Seed: %v", err)
	}
	if createdCode != "en" {
		t.Errorf("created %q, want %q", createdCode, "en")
	}
}

func TestSeed_NonEmptyTable_NoOp(t *testing.T) {
	useCase, localeRepo, _, _ := newTestUC()
	ctx := context.Background()

	localeRepo.FindAllFn = func(_ context.Context) ([]*entity.Locale, error) {
		return []*entity.Locale{{Code: "en"}}, nil
	}

	if err := useCase.Seed(ctx, []string{"en", "vi", "ja"}); err != nil {
		t.Fatalf("Seed: %v", err)
	}
}

func TestSeed_UsesKnownNames(t *testing.T) {
	useCase, localeRepo, _, _ := newTestUC()
	ctx := context.Background()

	localeRepo.FindAllFn = func(_ context.Context) ([]*entity.Locale, error) {
		return nil, nil
	}

	names := make(map[string]string)
	localeRepo.CreateFn = func(_ context.Context, loc *entity.Locale) error {
		names[loc.Code] = loc.Name
		return nil
	}

	if err := useCase.Seed(ctx, []string{"en", "vi"}); err != nil {
		t.Fatalf("Seed: %v", err)
	}
	if names["en"] != "English" {
		t.Errorf("en name = %q, want %q", names["en"], "English")
	}
	if names["vi"] != "Tiếng Việt" {
		t.Errorf("vi name = %q, want %q", names["vi"], "Tiếng Việt")
	}
}
