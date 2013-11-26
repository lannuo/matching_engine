package client

import (
	"strconv"
)

// TODO there is a defect here where the match price is half way between the bid/offer - when we reconcile the balances we are using the original bid/offer prices which leads to inconsistent balances
type balanceManager struct {
	Current   uint64 `json:"current"`
	Available uint64 `json:"available"`
}

func newBalanceManager() balanceManager {
	bal := uint64(100)
	return balanceManager{Current: bal, Available: bal}
}

func (bm *balanceManager) total(price uint64, amount uint32) uint64 {
	return price * uint64(amount)
}

func (bm *balanceManager) canBuy(price uint64, amount uint32) bool {
	return bm.Available >= bm.total(price, amount)
}

func (bm *balanceManager) submitBuy(price uint64, amount uint32) {
	bm.Available -= bm.total(price, amount)
}

func (bm *balanceManager) cancelBuy(price uint64, amount uint32) {
	bm.Available += bm.total(price, amount)
}

func (bm *balanceManager) completeBuy(price uint64, amount uint32) {
	bm.Current -= bm.total(price, amount)
}

func (bm *balanceManager) completeSell(price uint64, amount uint32) {
	total := bm.total(price, amount)
	bm.Current += total
	bm.Available += total
}

// TODO the naming of this hasn't worked out very well
type stockManager struct {
	StocksHeld   map[string]uint32 `json:"stocksHeld"`
	StocksToSell map[string]uint32 `json:"stocksToSell"`
}

func newStockManager() stockManager {
	sm := stockManager{}
	sm.StocksHeld = map[string]uint32{"1": 10, "2": 10, "3": 10}
	sm.StocksToSell = make(map[string]uint32)
	return sm
}

func (sm *stockManager) getKey(stockId uint32) string {
	return strconv.Itoa(int(stockId))
}

func (sm *stockManager) cleanup(stockKey string) {
	if sm.StocksToSell[stockKey] == 0 {
		delete(sm.StocksToSell, stockKey)
	}
	if sm.StocksHeld[stockKey] == 0 {
		delete(sm.StocksHeld, stockKey)
	}
}

func (sm *stockManager) held(stockKey string) uint32 {
	return sm.StocksHeld[stockKey]
}

func (sm *stockManager) addHeld(stockKey string, amount uint32) {
	held := sm.StocksHeld[stockKey]
	sm.StocksHeld[stockKey] = held + amount
}

func (sm *stockManager) toSell(stockKey string) uint32 {
	return sm.StocksToSell[stockKey]
}

func (sm *stockManager) addToSell(stockKey string, amount uint32) {
	toSell := sm.StocksToSell[stockKey]
	sm.StocksToSell[stockKey] = toSell + amount
}

func (sm *stockManager) canSell(stockId, amount uint32) bool {
	stockKey := sm.getKey(stockId)
	return sm.held(stockKey) >= amount
}

func (sm *stockManager) submitSell(stockId, amount uint32) {
	stockKey := sm.getKey(stockId)
	sm.addHeld(stockKey, -amount)
	sm.addToSell(stockKey, amount)
	// Don't clean up, we want the zeroed held stocks to remain
}

func (sm *stockManager) cancelSell(stockId, amount uint32) {
	stockKey := sm.getKey(stockId)
	sm.addHeld(stockKey, amount)
	sm.addToSell(stockKey, -amount)
	sm.cleanup(stockKey)
}

func (sm *stockManager) completeSell(stockId, amount uint32) {
	stockKey := sm.getKey(stockId)
	sm.addToSell(stockKey, -amount)
	sm.cleanup(stockKey)
}

func (sm *stockManager) completeBuy(stockId, amount uint32) {
	stockKey := sm.getKey(stockId)
	sm.addHeld(stockKey, amount)
}