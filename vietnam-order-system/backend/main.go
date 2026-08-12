package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type MenuItem struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Vietnamese  string  `json:"vietnamese"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	ImageURL    string  `json:"image_url"`
}

type OrderItem struct {
	ID         int     `json:"id"`
	OrderID    int     `json:"order_id"`
	MenuItemID int     `json:"menu_item_id"`
	Name       string  `json:"name"`
	Price      float64 `json:"price"`
	Quantity   int     `json:"quantity"`
}

type Order struct {
	ID          int         `json:"id"`
	TableNo     string      `json:"table_no"`
	Status      string      `json:"status"` // pending, paid
	TotalAmount float64     `json:"total_amount"`
	Items       []OrderItem `json:"items"`
	CreatedAt   time.Time   `json:"created_at"`
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./vietnam_order.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	initDB()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/menu", handleMenu)
	mux.HandleFunc("/api/orders", handleOrders)
	mux.HandleFunc("/api/orders/checkout", handleCheckout)

	corsMux := corsMiddleware(mux)

	log.Println("Server ready on port 8080...")
	if err := http.ListenAndServe(":8080", corsMux); err != nil {
		log.Fatal(err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Table-No")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func initDB() {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS menu_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			vietnamese TEXT,
			price REAL,
			category TEXT,
			description TEXT,
			image_url TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			table_no TEXT,
			status TEXT,
			total_amount REAL,
			created_at DATETIME
		);`,
		`CREATE TABLE IF NOT EXISTS order_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id INTEGER,
			menu_item_id INTEGER,
			quantity INTEGER,
			price REAL
		);`,
	}

	for _, stmt := range statements {
		_, err := db.Exec(stmt)
		if err != nil {
			log.Fatal(err)
		}
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM menu_items").Scan(&count)
	if count == 0 {
		db.Exec(`INSERT INTO menu_items (name, vietnamese, price, category, description, image_url) VALUES 
			('鶏肉のフォー', 'Phở Gà', 900, '麺類', 'あっさりした鶏出汁の定番フォー', 'https://images.unsplash.com/photo-1582878826629-29b7ad1cdc43?w=400'),
			('牛肉のフォー', 'Phở Bò', 980, '麺類', '旨味が凝縮された牛肉スープのフォー', 'https://images.unsplash.com/photo-1569718212165-3a8278d5f624?w=400'),
			('バインミー', 'Bánh Mì', 750, '軽食', '自家製パテと野菜たっぷりのベトナムサンドイッチ', 'https://images.unsplash.com/photo-1626844131082-256783844137?w=400'),
			('生春巻き', 'Gỏi Cuốn', 650, 'サイド', 'エビと豚肉、ハーブを包んだ新鮮な生春巻き（2本）', 'https://images.unsplash.com/photo-1534422298391-e4f8c172dddb?w=400'),
			('揚げ春巻き', 'Chả Giò', 700, 'サイド', 'カリッと揚げた具だくさんのベトナム風春巻き', 'https://images.unsplash.com/photo-1544025162-d76694265947?w=400'),
			('ベトナムコーヒー', 'Cà Phê Sữa Đá', 450, 'ドリンク', 'コンデンスミルク入りの濃厚で甘いアイスコーヒー', 'https://images.unsplash.com/photo-1517701604599-bb29b565090c?w=400')`)
	}
}

func handleMenu(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query("SELECT id, name, vietnamese, price, category, description, image_url FROM menu_items")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []MenuItem
	for rows.Next() {
		var item MenuItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Vietnamese, &item.Price, &item.Category, &item.Description, &item.ImageURL); err == nil {
			items = append(items, item)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func handleOrders(w http.ResponseWriter, r *http.Request) {
	tableNo := r.Header.Get("X-Table-No")
	if tableNo == "" {
		tableNo = "1"
	}

	if r.Method == "GET" {
		rows, err := db.Query("SELECT id, table_no, status, total_amount, created_at FROM orders WHERE table_no = ? ORDER BY created_at DESC", tableNo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var orders []Order
		for rows.Next() {
			var o Order
			rows.Scan(&o.ID, &o.TableNo, &o.Status, &o.TotalAmount, &o.CreatedAt)

			itemRows, _ := db.Query(`SELECT oi.id, oi.order_id, oi.menu_item_id, m.name, oi.price, oi.quantity 
				FROM order_items oi JOIN menu_items m ON oi.menu_item_id = m.id WHERE oi.order_id = ?`, o.ID)
			o.Items = []OrderItem{}
			for itemRows.Next() {
				var item OrderItem
				itemRows.Scan(&item.ID, &item.OrderID, &item.MenuItemID, &item.Name, &item.Price, &item.Quantity)
				o.Items = append(o.Items, item)
			}
			itemRows.Close()

			orders = append(orders, o)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(orders)

	} else if r.Method == "POST" {
		var req struct {
			Items []struct {
				MenuItemID int `json:"menu_item_id"`
				Quantity   int `json:"quantity"`
			} `json:"items"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if len(req.Items) == 0 {
			http.Error(w, "Order items empty", http.StatusBadRequest)
			return
		}

		var total float64 = 0
		type calcItem struct {
			menuID   int
			quantity int
			price    float64
		}
		var list []calcItem

		for _, item := range req.Items {
			var price float64
			err := db.QueryRow("SELECT price FROM menu_items WHERE id = ?", item.MenuItemID).Scan(&price)
			if err == nil {
				total += price * float64(item.Quantity)
				list = append(list, calcItem{menuID: item.MenuItemID, quantity: item.Quantity, price: price})
			}
		}

		res, err := db.Exec("INSERT INTO orders (table_no, status, total_amount, created_at) VALUES (?, 'pending', ?, ?)", tableNo, total, time.Now())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		orderID, _ := res.LastInsertId()
		for _, ci := range list {
			db.Exec("INSERT INTO order_items (order_id, menu_item_id, quantity, price) VALUES (?, ?, ?, ?)", orderID, ci.menuID, ci.quantity, ci.price)
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func handleCheckout(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tableNo := r.Header.Get("X-Table-No")
	if tableNo == "" {
		tableNo = "1"
	}

	_, err := db.Exec("UPDATE orders SET status = 'paid' WHERE table_no = ? AND status = 'pending'", tableNo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}