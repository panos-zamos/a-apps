package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/panos/a-apps/shared/auth"
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
		<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4">
			<div class="bg-white rounded-lg p-6 max-w-md w-full">
				<h3 class="text-xl font-bold mb-4">New Store</h3>
				<form hx-post="/stores" hx-target="#modal" class="space-y-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Store Name</label>
						<input 
							type="text" 
							name="name" 
							required
							placeholder="e.g., Supermarket, Pharmacy"
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Color</label>
						<input 
							type="color" 
							name="color" 
							value="#3B82F6"
							class="w-full h-10 rounded-lg"
						/>
					</div>
					<div class="flex gap-2">
						<button 
							type="submit" 
							class="flex-1 bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700"
						>
							Create
						</button>
						<button 
							type="button"
							onclick="this.closest('.fixed').remove()"
							class="flex-1 bg-gray-300 text-gray-700 px-4 py-2 rounded-lg hover:bg-gray-400"
						>
							Cancel
						</button>
					</div>
				</form>
			</div>
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
		<div class="flex items-center gap-2 p-2 hover:bg-gray-50 rounded">
			<input 
				type="checkbox"
				hx-post="/items/%d/toggle"
				hx-target="closest div"
				hx-swap="outerHTML"
				class="h-5 w-5 text-indigo-600 rounded"
			/>
			<span class="flex-1">%s</span>
			<span class="text-sm text-gray-500">%s</span>
			<button 
				hx-delete="/items/%d"
				hx-target="closest div"
				hx-swap="outerHTML"
				class="text-red-600 hover:text-red-800 text-sm"
			>
				✕
			</button>
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
	var checked int
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
	checkedClass := ""
	if checked == 1 {
		checkedAttr = "checked"
		checkedClass = "line-through text-gray-400"
	}

	// Return updated item HTML
	html := fmt.Sprintf(`
		<div class="flex items-center gap-2 p-2 hover:bg-gray-50 rounded">
			<input 
				type="checkbox"
				%s
				hx-post="/items/%s/toggle"
				hx-target="closest div"
				hx-swap="outerHTML"
				class="h-5 w-5 text-indigo-600 rounded"
			/>
			<span class="flex-1 %s">%s</span>
			<span class="text-sm text-gray-500 %s">%s</span>
			<button 
				hx-delete="/items/%s"
				hx-target="closest div"
				hx-swap="outerHTML"
				class="text-red-600 hover:text-red-800 text-sm"
			>
				✕
			</button>
		</div>
	`, checkedAttr, itemID, checkedClass, name, checkedClass, quantity, itemID)

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
	content := `<div id="stores-container" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">`

	if len(stores) == 0 {
		content += `
			<div class="col-span-full text-center py-12 text-gray-500">
				<p class="text-lg">No stores yet. Create your first shopping list!</p>
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
			<div class="bg-white rounded-lg shadow p-6 border-t-4" style="border-color: %s">
				<div class="flex justify-between items-start mb-4">
					<h3 class="text-xl font-bold text-gray-900">%s</h3>
					<button 
						hx-delete="/stores/%d" 
						hx-confirm="Delete this store and all items?"
						hx-target="#stores-container"
						hx-swap="outerHTML"
						class="text-red-600 hover:text-red-800"
					>
						🗑️
					</button>
				</div>
				
				<div class="mb-4 text-sm text-gray-600">
					%d items to buy
				</div>

				<form hx-post="/stores/%d/items" hx-target="#store-%d-items" hx-swap="beforeend" class="mb-4">
					<div class="flex gap-2">
						<input 
							type="text" 
							name="name" 
							placeholder="Add item..." 
							required
							class="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
						/>
						<input 
							type="text" 
							name="quantity" 
							placeholder="Qty" 
							class="w-20 px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
						/>
						<button type="submit" class="bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700">
							+
						</button>
					</div>
				</form>

				<div id="store-%d-items" class="space-y-2">
		`, store.Color, store.Name, store.ID, uncheckedCount, store.ID, store.ID, store.ID)

		for _, item := range store.Items {
			checkedClass := ""
			checkedAttr := ""
			if item.Checked {
				checkedClass = "line-through text-gray-400"
				checkedAttr = "checked"
			}

			content += fmt.Sprintf(`
				<div class="flex items-center gap-2 p-2 hover:bg-gray-50 rounded">
					<input 
						type="checkbox" 
						%s
						hx-post="/items/%d/toggle"
						hx-target="closest div"
						hx-swap="outerHTML"
						class="h-5 w-5 text-indigo-600 rounded"
					/>
					<span class="flex-1 %s">%s</span>
					<span class="text-sm text-gray-500 %s">%s</span>
					<button 
						hx-delete="/items/%d"
						hx-target="closest div"
						hx-swap="outerHTML"
						class="text-red-600 hover:text-red-800 text-sm"
					>
						✕
					</button>
				</div>
			`, checkedAttr, item.ID, checkedClass, item.Name, checkedClass, item.Quantity, item.ID)
		}

		content += `
				</div>
			</div>
		`
	}

	content += `</div>`
	return content
}
