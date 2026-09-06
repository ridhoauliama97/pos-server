# Graph Report - pos-server  (2026-09-06)

## Corpus Check
- 170 files · ~25,888 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 662 nodes · 1261 edges · 66 communities (17 shown, 48 thin omitted)
- Extraction: 96% EXTRACTED · 4% INFERRED · 0% AMBIGUOUS · INFERRED: 50 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- App Config Package
- AI
- Orm
- AdvancedReportTestSuite
- M20260906000011CreateTransactionItemsTable
- Transaction
- Goravel Framework
- M20260906000001CreateOutletsTable
- allowRoles
- M20260906000006CreatePaymentMethodsTable
- Crypt
- M20260906000003CreateCategoriesTable
- Gate Facade
- M20260906000004CreateProductsTable
- M20210101000001CreateJobsTable
- Mail Facade
- RateLimiter
- air.sh
- Cache
- UserController
- Session
- Seeder
- Telemetry
- Testing
- Validation
- Artisan
- Migrations
- Providers
- graphify.js
- AGENTS.md
- artisan script
- github.com/ridhoauliama97/pos-server
- ProductRequest
- DB
- App
- Auth
- Http
- Log
- M20260906000002CreateUsersTable
- Process
- Queue
- Schedule
- Storage
- View
- Schema
- M20260906000007CreateCustomersTable
- M20260906000008CreateStocksTable
- M20260906000009CreateStockMovementsTable
- M20260906000010CreateTransactionsTable
- M20260906000012CreateTransactionPaymentsTable
- Seeders
- ReportController
- StockService
- .Show
- TransactionTestSuite
- Hash
- CategoryRequest
- TransactionRequest
- github.com/goravel/framework/contracts/http.Context
- report_service.go
- RefreshRequest
- StockAdjustRequest
- github.com/goravel/framework/contracts/validation.Data
- ReceiptTestSuite
- UserRequest

## God Nodes (most connected - your core abstractions)
1. `Orm()` - 62 edges
2. `App()` - 31 edges
3. `Schema()` - 29 edges
4. `Transaction` - 24 edges
5. `Config()` - 19 edges
6. `Auth()` - 18 edges
7. `CustomerController` - 16 edges
8. `StockService` - 16 edges
9. `TransactionTestSuite` - 15 edges
10. `Api()` - 14 edges

## Surprising Connections (you probably didn't know these)
- `init()` --calls--> `Config()`  [EXTRACTED]
  config/ai.go → app/facades/config.go
- `init()` --calls--> `Config()`  [EXTRACTED]
  config/app.go → app/facades/config.go
- `init()` --calls--> `Config()`  [EXTRACTED]
  config/auth.go → app/facades/config.go
- `init()` --calls--> `Config()`  [EXTRACTED]
  config/cache.go → app/facades/config.go
- `init()` --calls--> `Config()`  [EXTRACTED]
  config/cors.go → app/facades/config.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Goravel Ecosystem** — goravel_framework, goravel_facades, artisan_console [EXTRACTED 0.90]

## Communities (66 total, 48 thin omitted)

### Community 0 - "App Config Package"
Cohesion: 0.06
Nodes (18): Config(), init(), init(), init(), init(), init(), init(), init() (+10 more)

### Community 2 - "Orm"
Cohesion: 0.07
Nodes (19): Gate(), Orm(), NewCategoryController(), NewCustomerController(), NewProductController(), serializeCategory(), NewStockController(), Category (+11 more)

### Community 3 - "AdvancedReportTestSuite"
Cohesion: 0.05
Nodes (18): AdvancedReportTestSuite, AuthTestSuite, CustomerTestSuite, ExampleTestSuite, IdempotencyTestSuite, ReportTestSuite, RoleMiddlewareTestSuite, github.com/goravel/framework/testing.TestCase (+10 more)

### Community 5 - "Transaction"
Cohesion: 0.07
Nodes (24): ProductVariant, Stock, StockMovement, Transaction, TransactionItem, User, NewCustomerService(), productKey() (+16 more)

### Community 6 - "Goravel Framework"
Cohesion: 0.29
Nodes (4): Goravel Development Skill, Artisan Console, Goravel Facades, Goravel Framework

### Community 8 - "allowRoles"
Cohesion: 0.11
Nodes (21): Lang(), RegisterCategory(), RegisterCustomer(), AllowOwners(), allowRoles(), and(), currentUser(), DenyOwnerRole() (+13 more)

### Community 21 - "Seeder"
Cohesion: 0.29
Nodes (3): Seeder(), github.com/goravel/framework/contracts/database/seeder.Facade, DatabaseSeeder

### Community 35 - "App"
Cohesion: 0.40
Nodes (3): App(), Event(), github.com/goravel/framework/contracts/event.Instance

### Community 36 - "Auth"
Cohesion: 0.06
Nodes (19): Auth(), Route(), NewAuthController(), NewTransactionController(), NewAuth(), NewRole(), NewTransactionService(), Boot() (+11 more)

### Community 45 - "Schema"
Cohesion: 0.29
Nodes (3): Schema(), github.com/goravel/framework/contracts/database/schema.Schema, M20260906000005CreateProductVariantsTable

### Community 52 - "ReportController"
Cohesion: 0.47
Nodes (3): NewReportController(), NewReportService(), ReportController

### Community 53 - "StockService"
Cohesion: 0.34
Nodes (4): NewStockService(), github.com/goravel/framework/contracts/database/orm.Query, StockChange, StockService

### Community 54 - ".Show"
Cohesion: 0.27
Nodes (4): NewReceiptController(), NewReceiptService(), ReceiptService, ReceiptController

### Community 56 - "Hash"
Cohesion: 0.17
Nodes (6): Hash(), isModelNotFound(), NewUserController(), github.com/goravel/framework/contracts/hash.Hash, UserSeeder, UserController

### Community 59 - "TransactionRequest"
Cohesion: 0.15
Nodes (5): BulkSyncRequest, BulkSyncTransactionRequest, TransactionItemRequest, TransactionPaymentRequest, TransactionRequest

### Community 61 - "github.com/goravel/framework/contracts/http.Context"
Cohesion: 0.19
Nodes (3): github.com/goravel/framework/contracts/http.Context, CustomerRequest, LoginRequest

### Community 62 - "report_service.go"
Cohesion: 0.24
Nodes (11): cashierName(), transactionDate(), CashierReport, CashierReportItem, DailyReport, DailyReportDay, ProductReport, ProductReportItem (+3 more)

## Knowledge Gaps
- **4 isolated node(s):** `air.sh script`, `github.com/ridhoauliama97/pos-server`, `Graphify`, `Artisan Console`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 183 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **48 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `App()` connect `App` to `App Config Package`, `AI`, `Orm`, `allowRoles`, `Crypt`, `Gate Facade`, `Mail Facade`, `RateLimiter`, `Cache`, `Session`, `Seeder`, `Telemetry`, `Testing`, `Validation`, `Artisan`, `DB`, `Auth`, `Http`, `Log`, `Process`, `Queue`, `Schedule`, `Storage`, `View`, `Schema`, `Hash`?**
  _High betweenness centrality (0.465) - this node is a cross-community bridge._
- **Why does `Orm()` connect `Orm` to `AdvancedReportTestSuite`, `App`, `ReceiptTestSuite`, `Transaction`, `Auth`, `StockService`, `.Show`, `TransactionTestSuite`, `Hash`, `report_service.go`?**
  _High betweenness centrality (0.439) - this node is a cross-community bridge._
- **Why does `Schema()` connect `Schema` to `App`, `M20260906000011CreateTransactionItemsTable`, `M20260906000002CreateUsersTable`, `M20260906000001CreateOutletsTable`, `M20260906000006CreatePaymentMethodsTable`, `M20260906000003CreateCategoriesTable`, `M20260906000004CreateProductsTable`, `M20210101000001CreateJobsTable`, `M20260906000007CreateCustomersTable`, `M20260906000008CreateStocksTable`, `M20260906000009CreateStockMovementsTable`, `M20260906000010CreateTransactionsTable`, `M20260906000012CreateTransactionPaymentsTable`?**
  _High betweenness centrality (0.184) - this node is a cross-community bridge._
- **Are the 29 inferred relationships involving `App()` (e.g. with `AI()` and `Artisan()`) actually correct?**
  _`App()` has 29 INFERRED edges - model-reasoned connections that need verification._
- **What connects `air.sh script`, `github.com/ridhoauliama97/pos-server`, `Graphify` to the rest of the system?**
  _4 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `App Config Package` be split into smaller, more focused modules?**
  _Cohesion score 0.05555555555555555 - nodes in this community are weakly interconnected._
- **Should `Orm` be split into smaller, more focused modules?**
  _Cohesion score 0.06526315789473684 - nodes in this community are weakly interconnected._