package laravel

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClassNameFromCollection(t *testing.T) {
	cases := map[string]string{"players": "Player", "categories": "Category", "purchase_orders": "PurchaseOrder"}
	for in, want := range cases {
		if got := ClassNameFromCollection(in); got != want {
			t.Fatalf("%s -> %s, want %s", in, got, want)
		}
	}
}

func TestInspectFindsMongoModels(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "artisan"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "composer.json"), []byte(`{"require":{"laravel/framework":"*"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	modelsDir := filepath.Join(root, "app", "Models")
	if err := os.MkdirAll(modelsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	php := `<?php
namespace App\Models;
use MongoDB\Laravel\Eloquent\Model;
class Developer extends Model {
    protected $table = 'developers';
    protected $connection = 'mongodb';
    public function team() { return $this->belongsTo(Team::class); }
}`
	if err := os.WriteFile(filepath.Join(modelsDir, "Developer.php"), []byte(php), 0o644); err != nil {
		t.Fatal(err)
	}
	info, err := Inspect(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(info.Models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(info.Models))
	}
	m := info.Models[0]
	if m.ClassName != "Developer" || m.Collection != "developers" || !m.MongoModel {
		t.Fatalf("unexpected model: %+v", m)
	}
	if len(m.Relationships) != 1 || m.Relationships[0] != "team(): belongsTo" {
		t.Fatalf("unexpected relationships: %+v", m.Relationships)
	}
}

func TestGeneratedModelUsesCurrentTableProperty(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "artisan"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "composer.json"), []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}
	target, err := GenerateModel(root, "developers", "Developer", `{"name":"Chris","tags":["PHP"]}`)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	if !strings.Contains(text, "protected $table = 'developers';") {
		t.Fatalf("missing table property: %s", text)
	}
	if strings.Contains(text, "'tags' => 'array'") {
		t.Fatalf("deprecated array cast generated: %s", text)
	}
}
