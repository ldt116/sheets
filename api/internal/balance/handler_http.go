package balance

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"api/internal/api"
	"api/model"
)

type Handler struct {
	svc    Service
	authMW echo.MiddlewareFunc
}

func (h *Handler) Bind(e *echo.Echo) {
	e.GET("/users", h.getUsers(), api.QueryParser())
	e.GET("/users/:id", h.getUser())

	g := e.Group("/balance")
	tx := g.Group("/transactions")
	tx.GET("", h.getTransactions(), api.QueryParser())
	tx.GET("/:id", h.getTransaction())
	tx.DELETE("/:id", h.deleteTransaction(), h.authMW)
	tx.PUT("/:id", h.updateTransaction(), h.authMW)
	tx.POST("", h.postTransaction(), h.authMW)
}

func (h *Handler) getUser() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		userID := c.Param("id")
		user, err := h.svc.User(ctx, userID)

		l := log.Ctx(ctx)
		if errors.Is(err, ErrNotFound) {
			return c.JSON(http.StatusNotFound, api.Response{Message: "User not found"})
		}
		if err != nil {
			l.Err(err).Str("userID", userID).Msg("FindUserByID failed")
			return c.JSON(http.StatusInternalServerError, api.Response{})
		}

		return c.JSON(http.StatusOK, api.Response{Success: true, Data: user})
	}
}

func (h *Handler) getTransaction() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		l := log.Ctx(ctx)

		txID := c.Param("id")
		tx, err := h.svc.Transaction(ctx, txID)
		if err != nil {
			if errors.Is(err, ErrTransactionNotFound) {
				return c.JSON(http.StatusNotFound, api.Response{Message: "Transaction not found"})
			}
			l.Err(err).Str("txID", txID).Msg("find transaction failed")
			return err
		}

		return c.JSON(http.StatusOK, api.Response{Success: true, Data: tx})
	}
}

func (h *Handler) getUsers() echo.HandlerFunc {
	type request struct {
	}

	return func(c echo.Context) error {
		var req request
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		ctx := c.Request().Context()
		l := log.Ctx(ctx)
		args := api.QueryFromContext(c)
		users, err := h.svc.Users(ctx, args)
		if err != nil {
			l.Err(err).Msg("list users failed")
			return err
		}

		return c.JSON(http.StatusOK, api.Response{Success: true, Data: users})
	}
}

func (h *Handler) getTransactions() echo.HandlerFunc {
	type request struct {
	}

	return func(c echo.Context) error {
		var req request
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		ctx := c.Request().Context()
		l := log.Ctx(ctx)
		args := api.QueryFromContext(c)
		txs, err := h.svc.Transactions(ctx, args)
		if err != nil {
			l.Err(err).Msg("list transactions failed")
			return err
		}

		return c.JSON(http.StatusOK, api.Response{Success: true, Data: txs})
	}
}

func (h *Handler) postTransaction() echo.HandlerFunc {
	type request struct {
		Transaction model.Transaction `json:"transaction" validate:"required"`
	}

	return func(ec echo.Context) error {
		c := ec.(*api.Context)
		var req request
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		l := log.Ctx(c.Ctx())
		tx, err := h.svc.AddTransaction(c.Ctx(), &req.Transaction, c.User())
		if err != nil {
			l.Err(err).Msg("add transaction failed")
			return err
		}

		return c.OK(tx)
	}
}

func (h *Handler) deleteTransaction() echo.HandlerFunc {
	return func(ec echo.Context) error {
		c := ec.(*api.Context)
		txID := c.Param("id")
		if len(txID) == 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid transaction ID")
		}

		l := log.Ctx(c.Ctx())
		if err := h.svc.DeleteTransaction(c.Ctx(), txID, c.User()); err != nil {
			if err == ErrTransactionNotFound {
				return echo.NewHTTPError(http.StatusNotFound, "Transaction not found")
			}
			l.Err(err).Msg("delete transaction failed")
			return err
		}

		return c.OK(nil)
	}
}

func (h *Handler) updateTransaction() echo.HandlerFunc {
	type request struct {
		Transaction TransactionDTO `json:"transaction" validate:"required"`
	}

	return func(ec echo.Context) error {
		c := ec.(*api.Context)
		var req request
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}
		txID := c.Param("id")

		l := log.Ctx(c.Ctx())
		if err := h.svc.UpdateTransaction(c.Ctx(), txID, &req.Transaction, c.User()); err != nil {
			l.Err(err).Msg("update transaction failed")
			return err
		}

		return c.OK(req.Transaction)
	}
}

func NewHandler(uc Service, authMW echo.MiddlewareFunc) *Handler {
	return &Handler{svc: uc, authMW: authMW}
}
