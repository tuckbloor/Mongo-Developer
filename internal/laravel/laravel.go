package laravel

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"go.mongodb.org/mongo-driver/v2/bson"
	"mongo-developer/internal/model"
)

type composerLock struct {
	Packages []struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"packages"`
}

func Inspect(path string) (model.LaravelProjectInfo, error) {
	clean, err := filepath.Abs(path)
	if err != nil {
		return model.LaravelProjectInfo{}, err
	}
	info := model.LaravelProjectInfo{Path: clean}
	info.ArtisanExists = fileExists(filepath.Join(clean, "artisan"))
	info.EnvExists = fileExists(filepath.Join(clean, ".env"))
	info.IsLaravel = info.ArtisanExists && fileExists(filepath.Join(clean, "composer.json"))
	lockPath := filepath.Join(clean, "composer.lock")
	if b, err := os.ReadFile(lockPath); err == nil {
		var lock composerLock
		if json.Unmarshal(b, &lock) == nil {
			for _, p := range lock.Packages {
				switch p.Name {
				case "laravel/framework":
					info.LaravelVersion = strings.TrimPrefix(p.Version, "v")
				case "mongodb/laravel-mongodb":
					info.MongoPackageInstalled = true
					info.MongoPackageVersion = strings.TrimPrefix(p.Version, "v")
				}
			}
		}
	}
	if b, err := os.ReadFile(filepath.Join(clean, ".env")); err == nil {
		text := string(b)
		info.MongoConfigured = strings.Contains(text, "DB_CONNECTION=mongodb") ||
			strings.Contains(text, "DB_URI=mongodb://") || strings.Contains(text, "DB_URI=mongodb+srv://")
	}
	info.Models = inspectModels(clean)
	return info, nil
}

func GenerateModel(projectPath, collection, className, sampleJSON string) (string, error) {
	info, err := Inspect(projectPath)
	if err != nil {
		return "", err
	}
	if !info.IsLaravel {
		return "", fmt.Errorf("selected folder is not a Laravel project")
	}
	collection = strings.TrimSpace(collection)
	if collection == "" {
		return "", fmt.Errorf("collection name is required")
	}
	if className == "" {
		className = ClassNameFromCollection(collection)
	}
	if !regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`).MatchString(className) {
		return "", fmt.Errorf("invalid PHP class name")
	}

	var sample bson.M
	if strings.TrimSpace(sampleJSON) != "" {
		if err := bson.UnmarshalExtJSON([]byte(sampleJSON), false, &sample); err != nil {
			return "", err
		}
	}
	fields := make([]string, 0, len(sample))
	casts := map[string]string{}
	for key, value := range sample {
		if key == "_id" {
			continue
		}
		fields = append(fields, key)
		switch value.(type) {
		case int32, int64:
			casts[key] = "integer"
		case float32, float64:
			casts[key] = "float"
		case bool:
			casts[key] = "boolean"
		case bson.DateTime:
			casts[key] = "datetime"
		}
	}
	sort.Strings(fields)
	modelDir := filepath.Join(projectPath, "app", "Models")
	if err := os.MkdirAll(modelDir, 0o755); err != nil {
		return "", err
	}
	target := filepath.Join(modelDir, className+".php")
	if fileExists(target) {
		return "", fmt.Errorf("model already exists: %s", target)
	}

	var b strings.Builder
	b.WriteString("<?php\n\nnamespace App\\Models;\n\n")
	b.WriteString("use MongoDB\\Laravel\\Eloquent\\Model;\n\n")
	b.WriteString("class " + className + " extends Model\n{\n")
	b.WriteString("    protected $connection = 'mongodb';\n\n")
	b.WriteString("    protected $table = '" + phpQuote(collection) + "';\n\n")
	b.WriteString("    protected $fillable = [\n")
	for _, f := range fields {
		b.WriteString("        '" + phpQuote(f) + "',\n")
	}
	b.WriteString("    ];\n")
	if len(casts) > 0 {
		b.WriteString("\n    protected $casts = [\n")
		keys := make([]string, 0, len(casts))
		for k := range casts {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			b.WriteString("        '" + phpQuote(k) + "' => '" + casts[k] + "',\n")
		}
		b.WriteString("    ];\n")
	}
	b.WriteString("}\n")
	if err := os.WriteFile(target, []byte(b.String()), 0o644); err != nil {
		return "", err
	}
	return target, nil
}

func inspectModels(projectPath string) []model.LaravelModelInfo {
	root := filepath.Join(projectPath, "app", "Models")
	models := []model.LaravelModelInfo{}
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".php") {
			return nil
		}
		b, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		text := string(b)
		classMatch := regexp.MustCompile(`(?m)\bclass\s+([A-Za-z_][A-Za-z0-9_]*)`).FindStringSubmatch(text)
		if len(classMatch) < 2 {
			return nil
		}
		className := classMatch[1]
		ns := ""
		if m := regexp.MustCompile(`(?m)^\s*namespace\s+([^;]+);`).FindStringSubmatch(text); len(m) > 1 {
			ns = strings.TrimSpace(m[1])
		}
		mongoModel := strings.Contains(text, `MongoDB\Laravel\Eloquent\Model`) || strings.Contains(text, `MongoDB\Laravel\Eloquent\DocumentModel`)
		collection := defaultCollectionForClass(className)
		legacy := false
		if m := regexp.MustCompile(`(?m)protected\s+\$table\s*=\s*['"]([^'"]+)['"]`).FindStringSubmatch(text); len(m) > 1 {
			collection = m[1]
		} else if m := regexp.MustCompile(`(?m)protected\s+\$collection\s*=\s*['"]([^'"]+)['"]`).FindStringSubmatch(text); len(m) > 1 {
			collection = m[1]
			legacy = true
		}
		connection := ""
		if m := regexp.MustCompile(`(?m)protected\s+\$connection\s*=\s*['"]([^'"]+)['"]`).FindStringSubmatch(text); len(m) > 1 {
			connection = m[1]
		}
		rels := []string{}
		relRe := regexp.MustCompile(`(?s)function\s+([A-Za-z_][A-Za-z0-9_]*)\s*\([^)]*\)[^{]*\{.{0,500}?\$this->(hasOne|hasMany|belongsTo|belongsToMany|embedsOne|embedsMany)\s*\(`)
		for _, m := range relRe.FindAllStringSubmatch(text, -1) {
			if len(m) > 2 {
				rels = append(rels, m[1]+"(): "+m[2])
			}
		}
		relPath, _ := filepath.Rel(projectPath, path)
		models = append(models, model.LaravelModelInfo{ClassName: className, Namespace: ns, Path: filepath.ToSlash(relPath), Collection: collection, Connection: connection, MongoModel: mongoModel, UsesLegacyField: legacy, Relationships: rels})
		return nil
	})
	sort.Slice(models, func(i, j int) bool { return models[i].ClassName < models[j].ClassName })
	return models
}

func defaultCollectionForClass(className string) string {
	var out strings.Builder
	runes := []rune(className)
	for i, r := range runes {
		if unicode.IsUpper(r) && i > 0 {
			out.WriteRune('_')
		}
		out.WriteRune(unicode.ToLower(r))
	}
	s := out.String()
	if strings.HasSuffix(s, "y") && len(s) > 1 {
		prev := s[len(s)-2]
		if !strings.ContainsRune("aeiou", rune(prev)) {
			return s[:len(s)-1] + "ies"
		}
	}
	if strings.HasSuffix(s, "s") {
		return s + "es"
	}
	return s + "s"
}

func ClassNameFromCollection(collection string) string {
	s := strings.TrimSpace(collection)
	if strings.HasSuffix(strings.ToLower(s), "ies") && len(s) > 3 {
		s = s[:len(s)-3] + "y"
	} else if strings.HasSuffix(strings.ToLower(s), "s") && len(s) > 1 {
		s = s[:len(s)-1]
	}
	parts := regexp.MustCompile(`[^A-Za-z0-9]+`).Split(s, -1)
	var out strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		r := []rune(p)
		r[0] = unicode.ToUpper(r[0])
		out.WriteString(string(r))
	}
	if out.Len() == 0 {
		return "MongoModel"
	}
	return out.String()
}

func Instructions(uri, db string) string {
	return fmt.Sprintf("composer require mongodb/laravel-mongodb\n\n.env:\nDB_CONNECTION=mongodb\nDB_URI=\"%s\"\nDB_DATABASE=%s\n", uri, db)
}

func fileExists(path string) bool { st, err := os.Stat(path); return err == nil && !st.IsDir() }
func phpQuote(s string) string    { return strings.ReplaceAll(s, "'", "\\'") }
