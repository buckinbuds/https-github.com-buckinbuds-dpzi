package data

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/harrybrwn/apizza/cmd/internal/out"
	"github.com/harrybrwn/apizza/dawg"
	"github.com/harrybrwn/apizza/pkg/cache"
	"github.com/harrybrwn/apizza/pkg/config"
	"github.com/harrybrwn/apizza/pkg/errs"
)

const (
	// OrderPrefix is the prefix added to user orders when stored in a database.
	OrderPrefix = "user_order_"

	// DataBaseName is the filename for the program's local storage.
	DataBaseName = "apizza.db"
)

// OpenDatabase make the default database.
func OpenDatabase() (*cache.DataBase, error) {
	dbPath := filepath.Join(config.Folder(), "cache", DataBaseName)
	return cache.GetDB(dbPath)
}

// ListOrders will return a list of orders stored in the database.
func ListOrders(db cache.MapDB) []string {
	all, _ := db.Map()
	names := make([]string, 0, len(all)) // its going to be at least as big
	for key := range all {
		if strings.Contains(key, OrderPrefix) {
			names = append(names, strings.Replace(key, OrderPrefix, "", -1))
		}
	}
	return names
}

// PrintOrders will print all the names of the saved user orders
func PrintOrders(db cache.MapDB, w io.Writer, verbose bool, color string) error {
	all, err := db.Map()
	if err != nil {
		return err
	}
	out.SetOutput(w)

	var (
		orders    = make([]string, 0, len(all)) // at least as big as all
		uOrders   []*dawg.Order
		tempOrder *dawg.Order
	)

	for k, v := range all {
		if strings.Contains(k, OrderPrefix) {
			name := strings.Replace(k, OrderPrefix, "", -1)
			orders = append(orders, name)

			if verbose {
				tempOrder = new(dawg.Order)
				if err = json.Unmarshal(v, tempOrder); err != nil {
					return err
				}
				tempOrder.OrderName = name
				uOrders = append(uOrders, tempOrder)
			}
		}
	}
	if len(orders) < 1 {
		fmt.Fprintln(w, "No orders saved.")
		return nil
	}

	var yesColor bool
	var endcolor = ""
	if color != "" {
		yesColor = true
		endcolor = "\033[0m"

	}
	fmt.Fprintf(w, "%sYour Orders%s:\n", color, endcolor)
	for i, o := range orders {
		if verbose {
			err = out.PrintOrder(uOrders[i], false, yesColor, false)
			if err != nil {
				return err
			}
		} else {
			fmt.Fprintln(w, " ", o)
		}
	}
	return nil
}

// GetOrder will get an order from a database.
func GetOrder(name string, db cache.Getter) (*dawg.Order, error) {
	raw, err := db.Get(OrderPrefix + name)
	if raw == nil {
		return nil, fmt.Errorf("cannot find order %s", name)
	}
	order := &dawg.Order{}
	order.Init()
	order.SetName(name)
	return order, errs.Pair(err, json.Unmarshal(raw, order))
}

// SaveOrder will save an order to a database.
//
// Also sends the order to the validation endpoint after saving it to the
// cache.Putter.
func SaveOrder(o *dawg.Order, w io.Writer, db cache.Putter) error {
	raw, err := json.Marshal(o)
	if err != nil {
		return err
	}
	err = db.Put(OrderPrefix+o.Name(), raw)
	if err == nil {
		fmt.Fprintln(w, "order successfully updated.")
	} else {
		return err
	}
	err = dawg.ValidateOrder(o)
	if dawg.IsFailure(err) {
		return err
	}
	return nil
}
