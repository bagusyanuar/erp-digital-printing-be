package http

import (
	"fmt"
	"math"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/product/delivery/http/dto"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/product/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type ProductHandler struct {
	productUsecase domain.ProductUsecase
}

func NewProductHandler(productUsecase domain.ProductUsecase) *ProductHandler {
	return &ProductHandler{productUsecase: productUsecase}
}

func (h *ProductHandler) Create(c fiber.Ctx) error {
	var req dto.CreateProductReq
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	product := &domain.Product{
		CategoryID: req.CategoryID,
		Name:       req.Name,
		SKU:        req.SKU,
		UOM:        req.UOM,
		BasePrice:  req.BasePrice,
	}

	// Map nested variants from request if provided
	if len(req.Variants) > 0 {
		product.Variants = make([]domain.ProductVariant, len(req.Variants))
		for i, v := range req.Variants {
			variant := domain.ProductVariant{
				VariantName:    v.VariantName,
				AdditionalCost: v.AdditionalCost,
				IsDefault:      false,
			}

			// Map attributes
			if len(v.Attributes) > 0 {
				variant.AttributeValues = make([]domain.ProductAttributeValue, len(v.Attributes))
				for j, a := range v.Attributes {
					variant.AttributeValues[j] = domain.ProductAttributeValue{
						AttributeID: a.AttributeID,
						Value:       a.Value,
					}
				}
			}

			// Map price tiers
			if len(v.PriceTiers) > 0 {
				variant.PriceTiers = make([]domain.PriceTier, len(v.PriceTiers))
				for j, t := range v.PriceTiers {
					variant.PriceTiers[j] = domain.PriceTier{
						CustomerLevelID: t.CustomerLevelID,
						MinQty:          t.MinQty,
						MaxQty:          t.MaxQty,
						PricePerUnit:    t.PricePerUnit,
					}
				}
			}

			product.Variants[i] = variant
		}
	}

	if err := h.productUsecase.Create(c.Context(), product); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to create product", err.Error())
	}

	// Map to response
	res := mapProductToRes(product)
	return response.Created(c, "Product created successfully", res)
}

func (h *ProductHandler) FindByID(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid product ID", err.Error())
	}

	product, err := h.productUsecase.FindByID(c.Context(), id)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "Product not found", err.Error())
	}

	res := mapProductToRes(product)
	return response.Success(c, "Product fetched successfully", res, nil)
}

func (h *ProductHandler) FindAll(c fiber.Ctx) error {
	var params request.PaginationParam
	if err := c.Bind().Query(&params); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid query parameters", err.Error())
	}

	search := c.Query("search", "")
	var categoryID *uuid.UUID
	if catStr := c.Query("category_id", ""); catStr != "" {
		if cid, err := uuid.Parse(catStr); err == nil {
			categoryID = &cid
		}
	}

	products, total, err := h.productUsecase.FindAll(c.Context(), params, search, categoryID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch products", err.Error())
	}

	resList := make([]dto.ProductRes, len(products))
	for i, p := range products {
		resList[i] = mapProductToRes(&p)
	}

	meta := response.Meta{
		Pagination: &response.Pagination{
			TotalItems:  total,
			TotalPages:  int(math.Ceil(float64(total) / float64(params.GetLimit()))),
			CurrentPage: params.GetPage(),
			Limit:       params.GetLimit(),
		},
	}

	return response.Success(c, "Products fetched successfully", resList, meta)
}

func (h *ProductHandler) Update(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid product ID", err.Error())
	}

	var req dto.UpdateProductReq
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	product, err := h.productUsecase.FindByID(c.Context(), id)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "Product not found", err.Error())
	}

	product.CategoryID = req.CategoryID
	product.Name = req.Name
	product.SKU = req.SKU
	product.UOM = req.UOM
	product.BasePrice = req.BasePrice

	// Map nested variants from request if provided
	if len(req.Variants) > 0 {
		product.Variants = make([]domain.ProductVariant, len(req.Variants))
		for i, v := range req.Variants {
			variant := domain.ProductVariant{
				VariantName:    v.VariantName,
				AdditionalCost: v.AdditionalCost,
				IsDefault:      false,
			}

			// Map attributes
			if len(v.Attributes) > 0 {
				variant.AttributeValues = make([]domain.ProductAttributeValue, len(v.Attributes))
				for j, a := range v.Attributes {
					variant.AttributeValues[j] = domain.ProductAttributeValue{
						AttributeID: a.AttributeID,
						Value:       a.Value,
					}
				}
			}

			// Map price tiers
			if len(v.PriceTiers) > 0 {
				variant.PriceTiers = make([]domain.PriceTier, len(v.PriceTiers))
				for j, t := range v.PriceTiers {
					variant.PriceTiers[j] = domain.PriceTier{
						CustomerLevelID: t.CustomerLevelID,
						MinQty:          t.MinQty,
						MaxQty:          t.MaxQty,
						PricePerUnit:    t.PricePerUnit,
					}
				}
			}

			product.Variants[i] = variant
		}
	} else {
		product.Variants = nil
	}

	if err := h.productUsecase.Update(c.Context(), product); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to update product", err.Error())
	}

	res := mapProductToRes(product)
	return response.Success(c, "Product updated successfully", res, nil)
}

func (h *ProductHandler) Delete(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid product ID", err.Error())
	}

	if err := h.productUsecase.Delete(c.Context(), id); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to delete product", err.Error())
	}

	return response.Success[any](c, "Product deleted successfully", nil, nil)
}

func (h *ProductHandler) CreateVariant(c fiber.Ctx) error {
	productIDStr := c.Params("product_id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid product ID", err.Error())
	}

	var req dto.CreateVariantReq
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	variant := &domain.ProductVariant{
		ProductID:      productID,
		VariantName:    req.VariantName,
		AdditionalCost: req.AdditionalCost,
		IsDefault:      false,
	}

	// Map nested attributes
	variant.AttributeValues = make([]domain.ProductAttributeValue, len(req.Attributes))
	for i, a := range req.Attributes {
		variant.AttributeValues[i] = domain.ProductAttributeValue{
			AttributeID: a.AttributeID,
			Value:       a.Value,
		}
	}

	// Map nested price tiers
	variant.PriceTiers = make([]domain.PriceTier, len(req.PriceTiers))
	for i, t := range req.PriceTiers {
		variant.PriceTiers[i] = domain.PriceTier{
			CustomerLevelID: t.CustomerLevelID,
			MinQty:          t.MinQty,
			MaxQty:          t.MaxQty,
			PricePerUnit:    t.PricePerUnit,
		}
	}

	if err := h.productUsecase.CreateVariant(c.Context(), variant); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to create variant with specs and price tiers", err.Error())
	}

	return response.Created(c, "Variant created successfully", mapVariantToRes(variant))
}

func (h *ProductHandler) CheckPrice(c fiber.Ctx) error {
	variantIDStr := c.Query("variant_id")
	customerLevelIDStr := c.Query("customer_level_id")
	qtyStr := c.Query("qty")

	if variantIDStr == "" || customerLevelIDStr == "" || qtyStr == "" {
		return response.Error(c, fiber.StatusBadRequest, "Missing required query parameters", "variant_id, customer_level_id, and qty are required")
	}

	variantID, err := uuid.Parse(variantIDStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid variant_id", err.Error())
	}

	customerLevelID, err := uuid.Parse(customerLevelIDStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid customer_level_id", err.Error())
	}

	qty := 0
	if _, err := fmt.Sscanf(qtyStr, "%d", &qty); err != nil || qty <= 0 {
		return response.Error(c, fiber.StatusBadRequest, "Invalid qty", "qty must be a positive integer")
	}

	result, err := h.productUsecase.CheckPrice(c.Context(), variantID, customerLevelID, qty)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "Price not found for given parameters", err.Error())
	}

	return response.Success(c, "Price checked successfully", result, nil)
}

// Helpers mapping
func mapProductToRes(p *domain.Product) dto.ProductRes {
	res := dto.ProductRes{
		ID:        p.ID,
		CategoryID: p.CategoryID,
		Name:      p.Name,
		SKU:       p.SKU,
		UOM:       p.UOM,
		BasePrice:  p.BasePrice,
		CreatedAt: p.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: p.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if p.Category.ID != uuid.Nil {
		res.CategoryName = p.Category.Name
	}

	if len(p.Variants) > 0 {
		res.Variants = make([]dto.ProductVariantRes, len(p.Variants))
		for i, v := range p.Variants {
			res.Variants[i] = mapVariantToRes(&v)
		}
	}

	return res
}

func mapVariantToRes(v *domain.ProductVariant) dto.ProductVariantRes {
	res := dto.ProductVariantRes{
		ID:             v.ID,
		VariantName:    v.VariantName,
		AdditionalCost: v.AdditionalCost,
		IsDefault:      v.IsDefault,
	}

	if len(v.AttributeValues) > 0 {
		res.AttributeValues = make([]dto.AttributeValueRes, len(v.AttributeValues))
		for i, av := range v.AttributeValues {
			res.AttributeValues[i] = dto.AttributeValueRes{
				ID:          av.ID,
				AttributeID: av.AttributeID,
				Value:       av.Value,
			}
			if av.Attribute.ID != uuid.Nil {
				res.AttributeValues[i].Name = av.Attribute.Name
				res.AttributeValues[i].Code = av.Attribute.Code
				res.AttributeValues[i].ValueType = av.Attribute.ValueType
			}
		}
	}

	if len(v.PriceTiers) > 0 {
		res.PriceTiers = make([]dto.PriceTierRes, len(v.PriceTiers))
		for i, pt := range v.PriceTiers {
			res.PriceTiers[i] = dto.PriceTierRes{
				ID:              pt.ID,
				CustomerLevelID: pt.CustomerLevelID,
				MinQty:          pt.MinQty,
				MaxQty:          pt.MaxQty,
				PricePerUnit:    pt.PricePerUnit,
			}
			if pt.CustomerLevel.ID != uuid.Nil {
				res.PriceTiers[i].CustomerLevelName = pt.CustomerLevel.Name
			}
		}
	}

	return res
}
