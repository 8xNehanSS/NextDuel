package questions

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"nextduel-svr/models"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func openBank(t *testing.T) *Bank {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "q.db")), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.Quest{}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	})
	return NewBank(db)
}

func TestSeedIsIdempotentAndServesValidQuestions(t *testing.T) {
	bank := openBank(t)
	ctx := context.Background()
	if err := bank.Seed(ctx); err != nil {
		t.Fatal(err)
	}
	if err := bank.Seed(ctx); err != nil {
		t.Fatalf("second seed failed: %v", err)
	}
	var count int64
	bank.db.Model(&models.Quest{}).Count(&count)
	if count < 20 {
		t.Fatalf("seeded %d questions, want at least 20", count)
	}

	qs, err := bank.Questions(ctx, 5)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, q := range qs {
		if seen[q.ID] {
			t.Fatalf("duplicate question %q in one draw", q.ID)
		}
		seen[q.ID] = true
		if len(q.Choices) != 4 || !q.HasChoice(q.Answer) {
			t.Fatalf("bad question: %+v", q)
		}
	}
}

func TestBankReportsShortage(t *testing.T) {
	bank := openBank(t)
	if _, err := bank.Questions(context.Background(), 3); !errors.Is(err, ErrNotEnoughQuestions) {
		t.Fatalf("err = %v, want ErrNotEnoughQuestions", err)
	}
}

func TestOpenTDBDecodesEntitiesAndRejectsRateLimit(t *testing.T) {
	code := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("amount") != "1" {
			t.Errorf("amount = %q", r.URL.Query().Get("amount"))
		}
		w.Header().Set("Content-Type", "application/json")
		if code != 0 {
			w.Write([]byte(`{"response_code":5,"results":[]}`))
			return
		}
		w.Write([]byte(`{"response_code":0,"results":[{"category":"Art","difficulty":"easy",
			"question":"Who painted &quot;Starry Night&quot;?","correct_answer":"Van Gogh",
			"incorrect_answers":["Monet","Dal&iacute;","Picasso"]}]}`))
	}))
	defer srv.Close()

	o := NewOpenTDB()
	o.Endpoint = srv.URL
	qs, err := o.Questions(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if qs[0].Text != `Who painted "Starry Night"?` || !qs[0].HasChoice("Dalí") {
		t.Fatalf("entities not decoded: %+v", qs[0])
	}

	code = 5
	if _, err := o.Questions(context.Background(), 1); err == nil {
		t.Fatal("expected error for rate-limited response")
	}
}

func TestFallbackUsesSecondaryAndCaches(t *testing.T) {
	bank := openBank(t)
	ctx := context.Background()
	if err := bank.Seed(ctx); err != nil {
		t.Fatal(err)
	}

	failing := providerFunc(func(context.Context, int) ([]Question, error) { return nil, errors.New("down") })
	f := &Fallback{Primary: failing, Secondary: bank}
	if qs, err := f.Questions(ctx, 3); err != nil || len(qs) != 3 {
		t.Fatalf("fallback: %v, %d questions", err, len(qs))
	}

	cached := 0
	fresh := providerFunc(func(_ context.Context, n int) ([]Question, error) {
		return []Question{buildQuestion("x", "X?", "a", []string{"b", "c", "d"}, "", "")}, nil
	})
	f = &Fallback{Primary: fresh, Secondary: bank, Cache: func(_ context.Context, qs []Question) { cached += len(qs) }}
	if _, err := f.Questions(ctx, 1); err != nil || cached != 1 {
		t.Fatalf("primary path: err=%v cached=%d", err, cached)
	}
}

type providerFunc func(context.Context, int) ([]Question, error)

func (f providerFunc) Questions(ctx context.Context, n int) ([]Question, error) { return f(ctx, n) }
