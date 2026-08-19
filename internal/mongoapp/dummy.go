package mongoapp

import (
	"context"
	"fmt"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"mongo-developer/internal/model"
)

const DummyDatabaseName = "dummy_test_data"

type dummyCollection struct {
	Name    string
	Docs    []any
	Indexes []mongo.IndexModel
}

func dt(year int, month time.Month, day, hour, minute int) time.Time {
	return time.Date(year, month, day, hour, minute, 0, 0, time.UTC)
}

func idx(keys bson.D, name string, unique bool) mongo.IndexModel {
	opts := options.Index().SetName(name)
	if unique {
		opts.SetUnique(true)
	}
	return mongo.IndexModel{Keys: keys, Options: opts}
}

func dummyCollections() []dummyCollection {
	departments := []any{
		bson.M{"_id": "dept_eng", "name": "Engineering", "cost_code": "ENG-100", "active": true},
		bson.M{"_id": "dept_prod", "name": "Product", "cost_code": "PROD-200", "active": true},
		bson.M{"_id": "dept_ops", "name": "Operations", "cost_code": "OPS-300", "active": true},
		bson.M{"_id": "dept_sales", "name": "Sales", "cost_code": "SAL-400", "active": true},
	}

	teams := []any{
		bson.M{"_id": "team_platform", "name": "Platform", "department_id": "dept_eng", "location": "London", "remote": true},
		bson.M{"_id": "team_web", "name": "Web Applications", "department_id": "dept_eng", "location": "Canterbury", "remote": true},
		bson.M{"_id": "team_data", "name": "Data", "department_id": "dept_eng", "location": "Manchester", "remote": true},
		bson.M{"_id": "team_product", "name": "Product Design", "department_id": "dept_prod", "location": "London", "remote": false},
		bson.M{"_id": "team_support", "name": "Customer Support", "department_id": "dept_ops", "location": "Leeds", "remote": true},
		bson.M{"_id": "team_sales", "name": "Enterprise Sales", "department_id": "dept_sales", "location": "London", "remote": false},
	}

	developerNames := []string{"Chris", "Aisha", "Ben", "Chloe", "Daniel", "Emma", "Farah", "George", "Hannah", "Imran", "Jess", "Kai", "Leah", "Marcus", "Nadia", "Owen", "Priya", "Ravi", "Sophie", "Tom"}
	teamIDs := []string{"team_web", "team_platform", "team_data", "team_web", "team_product", "team_platform", "team_data", "team_support", "team_web", "team_platform", "team_product", "team_data", "team_web", "team_support", "team_platform", "team_data", "team_sales", "team_web", "team_product", "team_platform"}
	roles := []string{"Senior Developer", "Backend Developer", "Data Engineer", "Frontend Developer", "Product Designer", "Platform Engineer", "Data Analyst", "Support Engineer", "Full Stack Developer", "DevOps Engineer", "UX Designer", "Database Engineer", "Laravel Developer", "Support Lead", "Security Engineer", "BI Developer", "Solutions Engineer", "Vue Developer", "Product Manager", "Go Developer"}
	developers := make([]any, 0, len(developerNames))
	for i, name := range developerNames {
		n := i + 1
		developers = append(developers, bson.M{
			"_id":              fmt.Sprintf("dev_%03d", n),
			"name":             name,
			"email":            fmt.Sprintf("%s%d@example.test", stringsLower(name), n),
			"team_id":          teamIDs[i],
			"role":             roles[i],
			"active":           i%7 != 0,
			"salary":           36000 + i*2750,
			"experience_years": 1 + (i % 12),
			"skills":           []string{"PHP", "Laravel", []string{"Vue", "Go", "MongoDB", "Docker"}[i%4]},
			"profile":          bson.M{"theme": []string{"dark", "light"}[i%2], "notifications": i%3 != 0, "timezone": "Europe/London"},
			"address":          bson.M{"city": []string{"Canterbury", "London", "Leeds", "Manchester"}[i%4], "country": "UK"},
			"joined_at":        dt(2021+i%5, time.Month(1+(i%12)), 2+(i%20), 9, 0),
			"last_login_at":    dt(2026, time.August, 1+(i%18), 8+(i%8), 15),
		})
	}

	customers := make([]any, 0, 18)
	for i := 1; i <= 18; i++ {
		customers = append(customers, bson.M{
			"_id":          fmt.Sprintf("cust_%03d", i),
			"company":      fmt.Sprintf("Example Customer %02d Ltd", i),
			"contact":      bson.M{"name": fmt.Sprintf("Contact %02d", i), "email": fmt.Sprintf("contact%02d@example.test", i), "phone": fmt.Sprintf("+44 20 7000 %04d", i)},
			"segment":      []string{"SMB", "Mid-market", "Enterprise"}[(i-1)%3],
			"active":       i%8 != 0,
			"credit_limit": 5000 + i*2500,
			"tags":         []string{"demo", []string{"retail", "manufacturing", "technology"}[(i-1)%3]},
			"address":      bson.M{"city": []string{"London", "Bristol", "Birmingham", "Edinburgh"}[(i-1)%4], "country": "UK", "postcode": fmt.Sprintf("D%02d 1AA", i)},
			"created_at":   dt(2024+(i%2), time.Month(1+(i%12)), 1+(i%25), 10, 30),
		})
	}

	suppliers := []any{}
	for i := 1; i <= 6; i++ {
		suppliers = append(suppliers, bson.M{"_id": fmt.Sprintf("sup_%02d", i), "name": fmt.Sprintf("Demo Supplier %02d", i), "rating": 3.5 + float64(i%3)*0.5, "active": i != 6, "lead_time_days": 2 + i})
	}

	categoryNames := []string{"Laptops", "Monitors", "Networking", "Storage", "Accessories", "Software"}
	categories := []any{}
	for i, name := range categoryNames {
		categories = append(categories, bson.M{"_id": fmt.Sprintf("cat_%02d", i+1), "name": name, "slug": stringsLower(name), "active": true})
	}

	products := make([]any, 0, 30)
	inventory := make([]any, 0, 30)
	for i := 1; i <= 30; i++ {
		cat := fmt.Sprintf("cat_%02d", 1+((i-1)%6))
		sup := fmt.Sprintf("sup_%02d", 1+((i-1)%6))
		price := 19.99 + float64(i*13)
		products = append(products, bson.M{
			"_id": fmt.Sprintf("prod_%03d", i), "sku": fmt.Sprintf("DEMO-%04d", i), "name": fmt.Sprintf("Demo Product %02d", i),
			"category_id": cat, "supplier_id": sup, "price": price, "cost": price * 0.62, "active": i%11 != 0,
			"tags":       []string{"demo", categoryNames[(i-1)%6], []string{"popular", "new", "clearance"}[(i-1)%3]},
			"specs":      bson.M{"weight_kg": 0.4 + float64(i%8)*0.35, "warranty_months": 12 * (1 + i%3), "colour": []string{"black", "silver", "white"}[i%3]},
			"created_at": dt(2025, time.Month(1+(i%12)), 1+(i%24), 12, 0),
		})
		inventory = append(inventory, bson.M{
			"_id": fmt.Sprintf("inv_%03d", i), "product_id": fmt.Sprintf("prod_%03d", i),
			"warehouse_id": []string{"wh_london", "wh_kent", "wh_manchester"}[(i-1)%3], "warehouse": []string{"London", "Kent", "Manchester"}[(i-1)%3], "quantity": (i * 7) % 53, "reorder_level": 10 + (i%5)*5,
			"bin": fmt.Sprintf("%c-%02d", 'A'+rune((i-1)%6), i), "updated_at": dt(2026, time.August, 1+(i%18), 7, 45),
		})
	}

	projects := make([]any, 0, 10)
	for i := 1; i <= 10; i++ {
		projects = append(projects, bson.M{
			"_id": fmt.Sprintf("proj_%02d", i), "name": fmt.Sprintf("Demo Project %02d", i),
			"customer_id": fmt.Sprintf("cust_%03d", 1+((i*2)%18)), "owner_id": fmt.Sprintf("dev_%03d", 1+((i*3)%20)),
			"status": []string{"planning", "active", "on_hold", "complete"}[(i-1)%4], "budget": 15000 + i*12000,
			"technologies": []string{"Laravel", "Vue", []string{"MongoDB", "Go", "Docker"}[i%3]},
			"milestones":   []bson.M{{"name": "Discovery", "complete": true}, {"name": "Build", "complete": i%3 == 0}, {"name": "Launch", "complete": i%4 == 0}},
			"started_at":   dt(2026, time.Month(1+(i%6)), 1+(i%20), 9, 0),
		})
	}

	projectMembers := []any{}
	memberID := 1
	for p := 1; p <= 10; p++ {
		for j := 0; j < 3; j++ {
			dev := 1 + ((p*2 + j*5) % 20)
			projectMembers = append(projectMembers, bson.M{"_id": fmt.Sprintf("pm_%03d", memberID), "project_id": fmt.Sprintf("proj_%02d", p), "developer_id": fmt.Sprintf("dev_%03d", dev), "role": []string{"Lead", "Developer", "Reviewer"}[j], "allocation_percent": []int{60, 40, 20}[j]})
			memberID++
		}
	}

	orders := []any{}
	orderItems := []any{}
	itemID := 1
	for i := 1; i <= 25; i++ {
		orderID := fmt.Sprintf("ord_%03d", i)
		status := []string{"new", "paid", "shipped", "complete", "cancelled"}[(i-1)%5]
		orders = append(orders, bson.M{
			"_id": orderID, "order_no": fmt.Sprintf("ORD-2026-%04d", i), "customer_id": fmt.Sprintf("cust_%03d", 1+((i-1)%18)),
			"status": status, "currency": "GBP", "discount_percent": (i % 4) * 5,
			"shipping":       bson.M{"method": []string{"standard", "express"}[i%2], "city": []string{"London", "Bristol", "Leeds"}[i%3], "tracking": fmt.Sprintf("TRACK%06d", i)},
			"status_history": []bson.M{{"status": "new", "at": dt(2026, time.August, 1+(i%15), 9, 0)}, {"status": status, "at": dt(2026, time.August, 2+(i%15), 14, 30)}},
			"created_at":     dt(2026, time.August, 1+(i%18), 9+(i%7), 10),
		})
		for j := 0; j < 3; j++ {
			prod := 1 + ((i*3 + j*7) % 30)
			qty := 1 + ((i + j) % 4)
			unit := 19.99 + float64(prod*13)
			orderItems = append(orderItems, bson.M{"_id": fmt.Sprintf("item_%03d", itemID), "order_id": orderID, "product_id": fmt.Sprintf("prod_%03d", prod), "quantity": qty, "unit_price": unit, "line_total": unit * float64(qty)})
			itemID++
		}
	}

	tickets := []any{}
	for i := 1; i <= 18; i++ {
		tickets = append(tickets, bson.M{
			"_id": fmt.Sprintf("ticket_%03d", i), "ticket_no": fmt.Sprintf("SUP-%04d", i), "customer_id": fmt.Sprintf("cust_%03d", 1+((i*4)%18)),
			"assigned_developer_id": fmt.Sprintf("dev_%03d", 1+((i*7)%20)), "priority": []string{"low", "normal", "high", "urgent"}[(i-1)%4],
			"status": []string{"open", "investigating", "waiting", "resolved"}[(i-1)%4], "subject": fmt.Sprintf("Demo support issue %02d", i),
			"labels":     []string{"demo", []string{"billing", "bug", "performance", "access"}[(i-1)%4]},
			"created_at": dt(2026, time.August, 1+(i%18), 8, 20), "resolved_at": func() any {
				if i%4 == 0 {
					return dt(2026, time.August, 2+(i%18), 16, 0)
				}
				return nil
			}(),
		})
	}

	events := []any{}
	for i := 1; i <= 24; i++ {
		events = append(events, bson.M{"_id": fmt.Sprintf("evt_%03d", i), "type": []string{"login", "deploy", "order_created", "ticket_updated"}[(i-1)%4], "actor_id": fmt.Sprintf("dev_%03d", 1+((i*5)%20)), "project_id": fmt.Sprintf("proj_%02d", 1+((i-1)%10)), "metadata": bson.M{"ip": fmt.Sprintf("192.0.2.%d", 10+i), "source": []string{"web", "api", "cli"}[i%3]}, "created_at": dt(2026, time.August, 1+(i%18), i%24, i%60)})
	}

	auditLogs := []any{}
	for i := 1; i <= 40; i++ {
		auditLogs = append(auditLogs, bson.M{"_id": fmt.Sprintf("audit_%03d", i), "action": []string{"create", "update", "delete", "login"}[(i-1)%4], "entity_type": []string{"product", "order", "customer", "ticket"}[(i-1)%4], "entity_id": fmt.Sprintf("entity_%03d", i), "user_id": fmt.Sprintf("dev_%03d", 1+((i*3)%20)), "changes": bson.M{"field": "status", "before": "old", "after": "new"}, "created_at": dt(2026, time.August, 1+(i%18), 6+(i%12), 5)})
	}

	settings := []any{
		bson.M{"_id": "app", "key": "app", "value": bson.M{"name": "Dummy Shop", "maintenance": false, "features": []string{"orders", "support", "projects"}}},
		bson.M{"_id": "notifications", "key": "notifications", "value": bson.M{"email": true, "sms": false, "digest_hour": 8}},
		bson.M{"_id": "limits", "key": "limits", "value": bson.M{"page_size": 50, "max_export": 5000}},
		bson.M{"_id": "nullable_example", "key": "nullable_example", "value": nil},
		bson.M{"_id": "array_example", "key": "array_example", "value": []any{"one", 2, true, nil, bson.M{"nested": "value"}}},
	}

	warehouses := []any{
		bson.M{"_id": "wh_london", "name": "London Warehouse", "city": "London", "capacity": 12000, "active": true},
		bson.M{"_id": "wh_kent", "name": "Kent Warehouse", "city": "Kent", "capacity": 8000, "active": true},
		bson.M{"_id": "wh_manchester", "name": "Manchester Warehouse", "city": "Manchester", "capacity": 9500, "active": true},
	}

	return []dummyCollection{
		{"departments", departments, []mongo.IndexModel{idx(bson.D{{Key: "name", Value: 1}}, "name_1", true)}},
		{"teams", teams, []mongo.IndexModel{idx(bson.D{{Key: "department_id", Value: 1}}, "department_id_1", false), idx(bson.D{{Key: "name", Value: 1}}, "name_1", true)}},
		{"developers", developers, []mongo.IndexModel{idx(bson.D{{Key: "email", Value: 1}}, "email_1", true), idx(bson.D{{Key: "team_id", Value: 1}}, "team_id_1", false), idx(bson.D{{Key: "active", Value: 1}, {Key: "role", Value: 1}}, "active_role", false)}},
		{"customers", customers, []mongo.IndexModel{idx(bson.D{{Key: "contact.email", Value: 1}}, "contact_email_1", true), idx(bson.D{{Key: "segment", Value: 1}}, "segment_1", false)}},
		{"suppliers", suppliers, []mongo.IndexModel{idx(bson.D{{Key: "name", Value: 1}}, "name_1", true)}},
		{"categories", categories, []mongo.IndexModel{idx(bson.D{{Key: "slug", Value: 1}}, "slug_1", true)}},
		{"products", products, []mongo.IndexModel{idx(bson.D{{Key: "sku", Value: 1}}, "sku_1", true), idx(bson.D{{Key: "category_id", Value: 1}}, "category_id_1", false), idx(bson.D{{Key: "supplier_id", Value: 1}}, "supplier_id_1", false), idx(bson.D{{Key: "price", Value: 1}}, "price_1", false)}},
		{"warehouses", warehouses, []mongo.IndexModel{idx(bson.D{{Key: "city", Value: 1}}, "city_1", false)}},
		{"inventory", inventory, []mongo.IndexModel{idx(bson.D{{Key: "product_id", Value: 1}, {Key: "warehouse_id", Value: 1}}, "product_warehouse", true), idx(bson.D{{Key: "quantity", Value: 1}}, "quantity_1", false)}},
		{"projects", projects, []mongo.IndexModel{idx(bson.D{{Key: "customer_id", Value: 1}}, "customer_id_1", false), idx(bson.D{{Key: "owner_id", Value: 1}}, "owner_id_1", false), idx(bson.D{{Key: "status", Value: 1}}, "status_1", false)}},
		{"project_members", projectMembers, []mongo.IndexModel{idx(bson.D{{Key: "project_id", Value: 1}, {Key: "developer_id", Value: 1}}, "project_developer", true), idx(bson.D{{Key: "developer_id", Value: 1}}, "developer_id_1", false)}},
		{"orders", orders, []mongo.IndexModel{idx(bson.D{{Key: "order_no", Value: 1}}, "order_no_1", true), idx(bson.D{{Key: "customer_id", Value: 1}}, "customer_id_1", false), idx(bson.D{{Key: "status", Value: 1}, {Key: "created_at", Value: -1}}, "status_created", false)}},
		{"order_items", orderItems, []mongo.IndexModel{idx(bson.D{{Key: "order_id", Value: 1}}, "order_id_1", false), idx(bson.D{{Key: "product_id", Value: 1}}, "product_id_1", false)}},
		{"support_tickets", tickets, []mongo.IndexModel{idx(bson.D{{Key: "ticket_no", Value: 1}}, "ticket_no_1", true), idx(bson.D{{Key: "customer_id", Value: 1}}, "customer_id_1", false), idx(bson.D{{Key: "assigned_developer_id", Value: 1}}, "assigned_developer_id_1", false), idx(bson.D{{Key: "status", Value: 1}, {Key: "priority", Value: 1}}, "status_priority", false)}},
		{"events", events, []mongo.IndexModel{idx(bson.D{{Key: "created_at", Value: -1}}, "created_at_desc", false), idx(bson.D{{Key: "type", Value: 1}}, "type_1", false)}},
		{"audit_logs", auditLogs, []mongo.IndexModel{idx(bson.D{{Key: "created_at", Value: -1}}, "created_at_desc", false), idx(bson.D{{Key: "user_id", Value: 1}}, "user_id_1", false)}},
		{"settings", settings, []mongo.IndexModel{idx(bson.D{{Key: "key", Value: 1}}, "key_1", true)}},
	}
}

func stringsLower(s string) string {
	b := make([]rune, 0, len(s))
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			r = r - 'A' + 'a'
		}
		if r == ' ' {
			r = '-'
		}
		b = append(b, r)
	}
	return string(b)
}

// EnsureDummyData installs the practice database only when it is absent.
func (m *Manager) EnsureDummyData(parent context.Context, id string) (model.DummyDataReport, error) {
	c, err := m.client(id)
	if err != nil {
		return model.DummyDataReport{}, err
	}
	ctx, cancel := context.WithTimeout(parent, 2*time.Minute)
	defer cancel()
	names, err := c.ListDatabaseNames(ctx, bson.D{{Key: "name", Value: DummyDatabaseName}})
	if err != nil {
		return model.DummyDataReport{}, err
	}
	if len(names) > 0 {
		report, err := m.DummyDataStatus(parent, id)
		report.Created = false
		return report, err
	}
	report, err := m.installDummyData(parent, id)
	if err == nil {
		report.Created = true
	}
	return report, err
}

// InstallDummyData drops only dummy_test_data and recreates the complete practice dataset.
func (m *Manager) InstallDummyData(parent context.Context, id string) (model.DummyDataReport, error) {
	report, err := m.installDummyData(parent, id)
	if err == nil {
		report.Created = true
	}
	return report, err
}

func (m *Manager) installDummyData(parent context.Context, id string) (model.DummyDataReport, error) {
	c, err := m.client(id)
	if err != nil {
		return model.DummyDataReport{}, err
	}
	ctx, cancel := context.WithTimeout(parent, 5*time.Minute)
	defer cancel()
	db := c.Database(DummyDatabaseName)
	if err := db.Drop(ctx); err != nil {
		return model.DummyDataReport{}, fmt.Errorf("reset %s: %w", DummyDatabaseName, err)
	}

	report := model.DummyDataReport{Database: DummyDatabaseName}
	for _, seed := range dummyCollections() {
		coll := db.Collection(seed.Name)
		if len(seed.Docs) > 0 {
			if _, err := coll.InsertMany(ctx, seed.Docs); err != nil {
				return report, fmt.Errorf("seed %s: %w", seed.Name, err)
			}
			report.Documents += len(seed.Docs)
		}
		if len(seed.Indexes) > 0 {
			names, err := coll.Indexes().CreateMany(ctx, seed.Indexes)
			if err != nil {
				return report, fmt.Errorf("index %s: %w", seed.Name, err)
			}
			report.Indexes += len(names)
		}
		report.Collections++
	}
	return report, nil
}

func (m *Manager) DummyDataStatus(parent context.Context, id string) (model.DummyDataReport, error) {
	c, err := m.client(id)
	if err != nil {
		return model.DummyDataReport{}, err
	}
	ctx, cancel := context.WithTimeout(parent, 2*time.Minute)
	defer cancel()
	names, err := c.ListDatabaseNames(ctx, bson.D{{Key: "name", Value: DummyDatabaseName}})
	if err != nil {
		return model.DummyDataReport{}, err
	}
	if len(names) == 0 {
		return model.DummyDataReport{Database: DummyDatabaseName}, nil
	}

	db := c.Database(DummyDatabaseName)
	collectionNames, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return model.DummyDataReport{}, err
	}
	sort.Strings(collectionNames)
	report := model.DummyDataReport{Database: DummyDatabaseName, Collections: len(collectionNames)}
	for _, name := range collectionNames {
		coll := db.Collection(name)
		count, err := coll.CountDocuments(ctx, bson.D{})
		if err != nil {
			return report, err
		}
		report.Documents += int(count)
		cur, err := coll.Indexes().List(ctx)
		if err == nil {
			// Include _id_ plus custom indexes in the status count.
			for cur.Next(ctx) {
				report.Indexes++
			}
			_ = cur.Close(ctx)
		}
	}
	return report, nil
}
