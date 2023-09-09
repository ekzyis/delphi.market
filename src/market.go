package main

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type Market struct {
	Id          int
	Description string
	Funding     int
	Active      bool
}

type Share struct {
	Id          string
	MarketId    int
	Description string
}

type Order struct {
	ShareId  string
	Side     string
	Price    int
	Quantity int
}

type OrderBookEntry struct {
	BuyQuantity  int
	BuyPrice     int
	SellPrice    int
	SellQuantity int
}

type MarketDataRequest struct {
	ShareId   string `json:"share_id"`
	OrderSide string `json:"side"`
	Quantity  int    `json:"quantity"`
}

func costFunction(b float64, q1 float64, q2 float64) float64 {
	// reference: http://blog.oddhead.com/2006/10/30/implementing-hansons-market-maker/
	return b * math.Log(math.Pow(math.E, q1/b)+math.Pow(math.E, q2/b))
}

// logarithmic market scoring rule (LMSR) market maker from Robin Hanson:
// https://mason.gmu.edu/~rhanson/mktscore.pdf1
func BinaryLMSR(invariant int, funding int, q1 int, q2 int, dq1 int) float64 {
	b := float64(funding)
	fq1 := float64(q1)
	fq2 := float64(q2)
	fdq1 := float64(dq1)
	return costFunction(b, fq1+fdq1, fq2) - costFunction(b, fq1, fq2)
}

func trades(c echo.Context) error {
	marketId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad Request")
	}
	var market Market
	if err = db.FetchMarket(int(marketId), &market); err == sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusNotFound, "Not Found")
	} else if err != nil {
		return err
	}
	var shares []Share
	if err = db.FetchShares(market.Id, &shares); err != nil {
		return err
	}
	data := map[string]any{
		"session":     c.Get("session"),
		"ENV":         ENV,
		"Id":          market.Id,
		"Description": market.Description,
		"Shares":      shares,
	}
	return c.Render(http.StatusOK, "bmarket_trade.html", data)
}

func orders(c echo.Context) error {
	marketId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad Request")
	}
	shareId := c.Param("sid")

	var market Market
	if err = db.FetchMarket(int(marketId), &market); err == sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusNotFound, "Not Found")
	} else if err != nil {
		return err
	}

	var shares []Share
	if err = db.FetchShares(market.Id, &shares); err != nil {
		return err
	}
	if shareId == "" {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/market/%d/%s", market.Id, shares[0].Id))
	}
	var orderBook []OrderBookEntry
	if err = db.FetchOrderBook(shareId, &orderBook); err != nil {
		return err
	}
	data := map[string]any{
		"session":     c.Get("session"),
		"ENV":         ENV,
		"Id":          market.Id,
		"Description": market.Description,
		"ShareId":     shareId,
		"Shares":      shares,
		"OrderBook":   orderBook,
	}
	return c.Render(http.StatusOK, "bmarket_order.html", data)
}
