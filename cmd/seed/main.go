package main

import (
	"log"

	rbacDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/rbac/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/shared/config"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/shared/database"
	userDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/user/domain"
	productDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/product/domain"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := database.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// 1. Seed Permissions
	permissions := []rbacDomain.Permission{
		{Name: "users:read", Resource: "users", Action: "read", Description: "View users"},
		{Name: "users:create", Resource: "users", Action: "create", Description: "Create users"},
		{Name: "users:update", Resource: "users", Action: "update", Description: "Update users"},
		{Name: "users:delete", Resource: "users", Action: "delete", Description: "Delete users"},
		{Name: "products:read", Resource: "products", Action: "read", Description: "View products"},
		{Name: "products:create", Resource: "products", Action: "create", Description: "Create products"},
	}

	for _, p := range permissions {
		db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoUpdates: clause.AssignmentColumns([]string{"resource", "action", "description"}),
		}).Create(&p)
	}

	// 2. Seed Roles
	roleNames := []string{"administrator", "admin", "designer"}
	roleDescriptions := map[string]string{
		"administrator": "Super Administrator - Full Access",
		"admin":         "Administrative Staff",
		"designer":      "Designer Staff",
	}

	for _, name := range roleNames {
		role := rbacDomain.Role{Name: name, Description: roleDescriptions[name]}
		db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoUpdates: clause.AssignmentColumns([]string{"description"}),
		}).Create(&role)
	}

	// Fetch roles to get IDs
	var superAdminRole, adminRole, designerRole rbacDomain.Role
	db.Where("name = ?", "administrator").First(&superAdminRole)
	db.Where("name = ?", "admin").First(&adminRole)
	db.Where("name = ?", "designer").First(&designerRole)

	// 3. Link Roles to Permissions
	// Admin: Manage users & products
	var adminPerms []rbacDomain.Permission
	db.Where("resource IN ?", []string{"users", "products"}).Find(&adminPerms)
	db.Model(&adminRole).Association("Permissions").Replace(adminPerms)

	// Designer: Read only
	var designerPerms []rbacDomain.Permission
	db.Where("action = ?", "read").Find(&designerPerms)
	db.Model(&designerRole).Association("Permissions").Replace(designerPerms)

	// 4. Seed Admin User
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("@Administrator1234"), bcrypt.DefaultCost)
	adminUser := userDomain.User{
		Username: "administrator",
		Password: string(hashedPassword),
	}

	db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "username"}},
		DoUpdates: clause.AssignmentColumns([]string{"password"}),
	}).Create(&adminUser)

	// Fetch user to get ID
	db.Where("username = ?", "administrator").First(&adminUser)

	// 5. Link User to Role
	db.Model(&adminUser).Association("Roles").Replace(&superAdminRole)

	// 6. Sync to Casbin
	syncCasbin(db)

	// 7. Seed Customer Levels
	customerLevels := []productDomain.CustomerLevel{
		{
			ID:                 uuid.MustParse("b3c8f3a3-b26a-4638-b7f2-841a54774844"),
			Name:               "End User",
			DiscountPercentage: 0,
		},
		{
			ID:                 uuid.MustParse("d2c67ef8-82e4-4d8b-968b-5a1e2f5b6154"),
			Name:               "Reseller",
			DiscountPercentage: 0,
		},
	}

	for _, cl := range customerLevels {
		db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "discount_percentage"}),
		}).Create(&cl)
	}

	log.Println("✅ Seeding RBAC & Customer Levels completed: End User & Reseller created")
}

func syncCasbin(db *gorm.DB) {
	db.Exec("TRUNCATE TABLE casbin_rule")

	// Special Rule for administrator (Wildcard)
	db.Table("casbin_rule").Create(map[string]any{
		"ptype": "p",
		"v0":    "administrator",
		"v1":    "*",
		"v2":    "*",
	})

	// Sync Policies (p) for other roles
	var roles []rbacDomain.Role
	db.Preload("Permissions").Where("name != ?", "administrator").Find(&roles)
	for _, r := range roles {
		for _, p := range r.Permissions {
			db.Table("casbin_rule").Create(map[string]any{
				"ptype": "p",
				"v0":    r.Name,
				"v1":    p.Resource,
				"v2":    p.Action,
			})
		}
	}

	// Sync Grouping (g)
	var users []userDomain.User
	db.Preload("Roles").Find(&users)
	for _, u := range users {
		for _, r := range u.Roles {
			db.Table("casbin_rule").Create(map[string]any{
				"ptype": "g",
				"v0":    u.ID.String(),
				"v1":    r.Name,
			})
		}
	}
}
