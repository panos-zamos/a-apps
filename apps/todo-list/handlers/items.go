package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/panos-zamos/a-apps/shared/auth"
)

// CreateStore handles creating a new store
func (h *Handler) CreateStore(w http.ResponseWriter, r *http.Request) {
	username, _ := auth.GetUsername(r)
	name := r.FormValue("name")
	color := r.FormValue("color")

	if name == "" {
		http.Error(w, "Store name required", http.StatusBadRequest)
		return
	}

	if color == "" {
		color = "#3B82F6"
	}

	_, err := h.DB.Exec(`
		INSERT INTO stores (name, username, color) 
		VALUES (?, ?, ?)
	`, name, username, color)

	if err != nil {
		http.Error(w, "Failed to create store", http.StatusInternalServerError)
		return
	}

	// Redirect to home to reload
	w.Header().Set("HX-Redirect", "/")
}

// DeleteStore handles deleting a store
func (h *Handler) DeleteStore(w http.ResponseWriter, r *http.Request) {
	username, _ := auth.GetUsername(r)
	storeID := chi.URLParam(r, "id")

	_, err := h.DB.Exec(`
		DELETE FROM stores 
		WHERE id = ? AND username = ?
	`, storeID, username)

	if err != nil {
		http.Error(w, "Failed to delete store", http.StatusInternalServerError)
		return
	}

	// Return updated stores list
	stores, _ := h.getStores(username)
	w.Write([]byte(h.storesGrid(stores)))
}

// NewStoreForm returns the form for creating a new store
func (h *Handler) NewStoreForm(w http.ResponseWriter, r *http.Request) {
	html := `
		<div class="panel">
			<h2>New Store</h2>
			<form hx-post="/stores" hx-target="#modal" class="mt-md">
				<label>Store Name</label>
				<input type="text" name="name" required placeholder="e.g., Supermarket, Pharmacy">
				<div class="mt-md">
					<label>Color</label>
					<input type="color" name="color" value="#3B82F6">
				</div>
				<div class="row mt-md">
					<button type="submit" class="primary">Create</button>
					<button type="button" onclick="this.closest('.panel').remove()">Cancel</button>
				</div>
			</form>
		</div>
	`
	w.Write([]byte(html))
}

// CreateItem handles adding an item to a store
func (h *Handler) CreateItem(w http.ResponseWriter, r *http.Request) {
	username, _ := auth.GetUsername(r)
	storeID := chi.URLParam(r, "storeID")
	name := r.FormValue("name")
	quantity := r.FormValue("quantity")

	if name == "" {
		http.Error(w, "Item name required", http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec(`
		INSERT INTO items (store_id, name, quantity, username) 
		VALUES (?, ?, ?, ?)
	`, storeID, name, quantity, username)

	if err != nil {
		http.Error(w, "Failed to create item", http.StatusInternalServerError)
		return
	}

	itemID, _ := result.LastInsertId()

	// Return the new item HTML
	html := fmt.Sprintf(`
		<div class="row space-between mb-sm">
			<div class="row">
				<input type="checkbox" hx-post="/items/%d/toggle" hx-target="closest .space-between" hx-swap="outerHTML">
				<span>%s</span>
				<span class="muted">%s</span>
			</div>
			<button class="danger" hx-delete="/items/%d" hx-target="closest .space-between" hx-swap="outerHTML">Remove</button>
		</div>
	`, itemID, name, quantity, itemID)

	w.Write([]byte(html))
}

// ToggleItem handles checking/unchecking an item
func (h *Handler) ToggleItem(w http.ResponseWriter, r *http.Request) {
	username, _ := auth.GetUsername(r)
	itemID := chi.URLParam(r, "id")

	// Toggle the checked status
	_, err := h.DB.Exec(`
		UPDATE items 
		SET checked = NOT checked 
		WHERE id = ? AND username = ?
	`, itemID, username)

	if err != nil {
		http.Error(w, "Failed to toggle item", http.StatusInternalServerError)
		return
	}

	// Get updated item
	var name, quantity string
	var checked bool
	err = h.DB.QueryRow(`
		SELECT name, quantity, checked 
		FROM items 
		WHERE id = ?
	`, itemID).Scan(&name, &quantity, &checked)

	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	checkedAttr := ""
	itemName := name
	itemQuantity := quantity
	if checked {
		checkedAttr = "checked"
		itemName = fmt.Sprintf("<del>%s</del>", name)
		itemQuantity = fmt.Sprintf("<del>%s</del>", quantity)
	}

	// Return updated item HTML
	html := fmt.Sprintf(`
		<div class="row space-between mb-sm">
			<div class="row">
				<input type="checkbox" %s hx-post="/items/%s/toggle" hx-target="closest .space-between" hx-swap="outerHTML">
				<span>%s</span>
				<span class="muted">%s</span>
			</div>
			<button class="danger" hx-delete="/items/%s" hx-target="closest .space-between" hx-swap="outerHTML">Remove</button>
		</div>
	`, checkedAttr, itemID, itemName, itemQuantity, itemID)

	w.Write([]byte(html))
}

// DeleteItem handles deleting an item
func (h *Handler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	username, _ := auth.GetUsername(r)
	itemID := chi.URLParam(r, "id")

	_, err := h.DB.Exec(`
		DELETE FROM items 
		WHERE id = ? AND username = ?
	`, itemID, username)

	if err != nil {
		http.Error(w, "Failed to delete item", http.StatusInternalServerError)
		return
	}

	// Return empty (item will be removed from DOM)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) storesGrid(stores []Store) string {
	content := `<div id="stores-container" class="card-list">`

	if len(stores) == 0 {
		content += `
			<div class="panel center">
				<p class="muted">No stores yet. Create your first shopping list!</p>
			</div>
		`
	}

	for _, store := range stores {
		uncheckedCount := 0
		for _, item := range store.Items {
			if !item.Checked {
				uncheckedCount++
			}
		}

		content += fmt.Sprintf(`
			<article class="card">
				<div class="row space-between mb-sm">
					<span class="card-name">%s</span>
					<button class="danger" hx-delete="/stores/%d" hx-confirm="Delete this store and all items?" hx-target="#stores-container" hx-swap="outerHTML">Delete</button>
				</div>

				<p class="muted mb-sm">%d items to buy</p>

				<div id="store-%d-items">
		`, store.Name, store.ID, uncheckedCount, store.ID)

		for _, item := range store.Items {
			checkedAttr := ""
			itemName := item.Name
			itemQuantity := item.Quantity
			if item.Checked {
				checkedAttr = "checked"
				itemName = fmt.Sprintf("<del>%s</del>", item.Name)
				itemQuantity = fmt.Sprintf("<del>%s</del>", item.Quantity)
			}

			content += fmt.Sprintf(`
				<div class="row space-between mb-sm">
					<div class="row">
						<input type="checkbox" %s hx-post="/items/%d/toggle" hx-target="closest .space-between" hx-swap="outerHTML">
						<span>%s</span>
						<span class="muted">%s</span>
					</div>
					<button class="danger" hx-delete="/items/%d" hx-target="closest .space-between" hx-swap="outerHTML">Remove</button>
				</div>
			`, checkedAttr, item.ID, itemName, itemQuantity, item.ID)
		}

		content += fmt.Sprintf(`
				</div>

				<form hx-post="/stores/%d/items" hx-target="#store-%d-items" hx-swap="beforeend" class="mt-md">
					<label>Item</label>
					<input type="text" name="name" placeholder="Add item..." required>
					<div class="mt-sm">
						<label>Quantity</label>
						<input type="text" name="quantity" placeholder="Qty (optional)">
					</div>
					<div class="mt-sm">
						<button type="submit" class="primary">Add Item</button>
					</div>
				</form>
			</article>
		`, store.ID, store.ID)
	}

	content += `</div>`
	return content
}
