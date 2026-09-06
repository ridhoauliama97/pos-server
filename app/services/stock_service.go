package services

import (
	"errors"

	"gorm.io/gorm"

	contractsorm "github.com/goravel/framework/contracts/database/orm"

	"github.com/ridhoauliama97/pos-server/app/facades"
	"github.com/ridhoauliama97/pos-server/app/models"
)

const (
	StockMovementIn         = "in"
	StockMovementOut        = "out"
	StockMovementAdjustment = "adjustment"
	StockMovementSale       = "sale"
	StockMovementReturn     = "return"
)

var ErrInsufficientStock = errors.New("insufficient stock")

// StockChange describes a single stock mutation on one product variant.
type StockChange struct {
	OutletId         int64
	ProductVariantId int64
	Quantity         float64
	ReferenceType    *string
	ReferenceId      *int64
	Note             *string
	CreatedBy        *int64
}

type StockService struct{}

func NewStockService() *StockService {
	return &StockService{}
}

func (s *StockService) IncrementStock(change StockChange) error {
	return s.incrementStock(facades.Orm().Query(), change, StockMovementIn)
}

func (s *StockService) IncrementStockWithTx(tx contractsorm.Query, change StockChange) error {
	return s.incrementStock(tx, change, StockMovementIn)
}

func (s *StockService) ReturnStock(change StockChange) error {
	return s.incrementStock(facades.Orm().Query(), change, StockMovementReturn)
}

func (s *StockService) ReturnStockWithTx(tx contractsorm.Query, change StockChange) error {
	return s.incrementStock(tx, change, StockMovementReturn)
}

func (s *StockService) DecrementStock(change StockChange) error {
	return s.decrementStock(facades.Orm().Query(), change, StockMovementSale)
}

func (s *StockService) DecrementStockWithTx(tx contractsorm.Query, change StockChange) error {
	return s.decrementStock(tx, change, StockMovementSale)
}

func (s *StockService) AdjustStock(change StockChange) error {
	tx, err := facades.Orm().Query().BeginTransaction()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.adjustStock(tx, change); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *StockService) AdjustStockWithTx(tx contractsorm.Query, change StockChange) error {
	return s.adjustStock(tx, change)
}

func (s *StockService) incrementStock(tx contractsorm.Query, change StockChange, movementType string) error {
	if change.Quantity <= 0 {
		return errors.New("quantity must be greater than zero")
	}

	if err := s.ensureStockRow(tx, change.OutletId, change.ProductVariantId); err != nil {
		return err
	}

	if _, err := tx.Model(&models.Stock{}).
		Where("outlet_id = ? AND product_variant_id = ?", change.OutletId, change.ProductVariantId).
		Update("quantity", gorm.Expr("quantity + ?", change.Quantity)); err != nil {
		return err
	}

	return s.recordMovement(tx, change, movementType)
}

func (s *StockService) decrementStock(tx contractsorm.Query, change StockChange, movementType string) error {
	if change.Quantity <= 0 {
		return errors.New("quantity must be greater than zero")
	}

	if err := s.ensureStockRow(tx, change.OutletId, change.ProductVariantId); err != nil {
		return err
	}

	rows, err := tx.Model(&models.Stock{}).
		Where("outlet_id = ? AND product_variant_id = ? AND quantity >= ?", change.OutletId, change.ProductVariantId, change.Quantity).
		Update("quantity", gorm.Expr("quantity - ?", change.Quantity))
	if err != nil {
		return err
	}
	if rows.RowsAffected == 0 {
		return ErrInsufficientStock
	}

	return s.recordMovement(tx, change, movementType)
}

func (s *StockService) adjustStock(tx contractsorm.Query, change StockChange) error {
	if change.Quantity < 0 {
		return errors.New("quantity must not be negative")
	}

	if err := s.ensureStockRow(tx, change.OutletId, change.ProductVariantId); err != nil {
		return err
	}

	var stock models.Stock
	if err := tx.Where("outlet_id = ? AND product_variant_id = ?", change.OutletId, change.ProductVariantId).LockForUpdate().First(&stock); err != nil {
		return err
	}

	delta := change.Quantity - stock.Quantity
	if delta != 0 {
		if _, err := tx.Model(&models.Stock{}).
			Where("outlet_id = ? AND product_variant_id = ?", change.OutletId, change.ProductVariantId).
			Update("quantity", gorm.Expr("quantity + ?", delta)); err != nil {
			return err
		}
	}

	return s.recordMovement(tx, change, StockMovementAdjustment)
}

func (s *StockService) ensureStockRow(tx contractsorm.Query, outletId, productVariantId int64) error {
	var stock models.Stock
	err := tx.UpdateOrCreate(&stock,
		map[string]any{
			"outlet_id":          outletId,
			"product_variant_id": productVariantId,
		},
		map[string]any{
			"outlet_id":          outletId,
			"product_variant_id": productVariantId,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *StockService) recordMovement(tx contractsorm.Query, change StockChange, movementType string) error {
	movement := models.StockMovement{
		OutletId:         change.OutletId,
		ProductVariantId: change.ProductVariantId,
		Type:             movementType,
		Quantity:         change.Quantity,
		ReferenceType:    change.ReferenceType,
		ReferenceId:      change.ReferenceId,
		Note:             change.Note,
		CreatedBy:        change.CreatedBy,
	}

	return tx.Create(&movement)
}
