package app

import (
	"fmt"
	"github.com/adshao/go-binance"
	"time"
)

type Logic struct {
	b *MyBinance
}

func NewLogic(b *MyBinance) *Logic {
	return &Logic{
		b: b,
	}
}

type Balance struct {
	usd    string
	asset  string
	free   string
	locked string
}

type Status struct {
	total	 string
	rate     string
	balances []*Balance
}

func (l *Logic) CommandStatus() (*Status, error) {
	rate, err := l.b.GetRate()
	if err != nil {
		return nil, err
	}
	allBalances, err := l.b.AccountBalance()
	if err != nil {
		return nil, err
	}

	var balances []*Balance
	var total float64
	for _, bal := range allBalances {
		if isEmptyBalance(bal.Free) && isEmptyBalance(bal.Locked) {
			continue
		}

		balUSD, err := l.b.BalanceToUSD(&bal)
		if err != nil {
			return &Status{}, err
		}
		total += balUSD
		resBal := &Balance{
			   usd:    float64ToStr(balUSD),
			   asset:  bal.Asset,
			   free:   bal.Free,
			   locked: bal.Locked,
		}
		balances = append(balances, resBal)
	}

	res := &Status{
		total:	  float64ToStr(total),
		rate:     rate,
		balances: balances,
	}
	return res, err
}

const sleepDur = time.Duration(2) * time.Second

func (l *Logic) CommandBuy(s *Sender) {
	for i := 0; i < 5; i++ {
		orderNew, err := l.b.BuyAll()
		if err != nil {
			s.Send(errorMessage(err))
			return
		}
		s.Send(startMessage(&OrderNew{orderNew}))
		time.Sleep(sleepDur)
		order, err := l.b.GetOrder(orderNew.OrderID)
		if err != nil {
			s.Send(errorMessage(err))
			return
		}
		s.Send(buyStatusMessage(&OrderExist{order}))
		if order.Status == binance.OrderStatusTypeFilled {
			congratsMessage(i)
			return
		}
		err = l.b.CancelOrder(order.OrderID)
		if err != nil {
			s.Send(errorMessage(err))
			return
		}
	}
}

//--------------------------------------TEMPLATES FOR SENDER----------------------------------------------
func errorMessage(err error) string {
	return fmt.Sprintf("Error while Buy:\n\n%s", err)
}

func congratsMessage(i int) string {
	return fmt.Sprintf("Congratulations! Order filled in %v iterations!", i)
}

func startMessage(order Order) string {
	return fmt.Sprintf("A %v BTC/USDT order was placed with price = %v.\nWaiting for 2 seconds..", order.Side(), order.Price())
}

func buyStatusMessage(order Order) string {
	return fmt.Sprintf("Done %v / %v\nStatus: %v", order.ExecutedQuantity(), order.OrigQuantity(), order.Status())
}
//-------------------------------------------------------------------------------

//TEST COMMANDS
func (l *Logic) TestCommandBuy(s *Sender) {
	for i := 0; i < 5; i++ {
		err := l.b.TestBuyAll()
		if err != nil {
			s.Send(errorMessage(err))
			return
		}
		s.Send("START")
		time.Sleep(sleepDur)
		if err != nil {
			s.Send(errorMessage(err))
			return
		}
		s.Send("KekWait")
		if err != nil {
			s.Send(errorMessage(err))
			return
		}
	}
}
