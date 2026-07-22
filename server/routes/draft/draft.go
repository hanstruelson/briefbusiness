package draft

import (
	"appeals/db"
	"appeals/utilities"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Draft struct {
	Id        int       `json:"id"`
	Title     string    `json:"title"`
	Markdown  string    `json:"markdown"`
	UpdatedAt time.Time `json:"updated_at"`
	UserId    int       `json:"user_id"`
}

type DraftBackupInfo struct {
	Id      int       `json:"id"`
	DraftId int       `json:"draft_id"`
	Title   string    `json:"title"`
	SavedAt time.Time `json:"saved_at"`
}

type DraftDetail struct {
	Draft
	Backups []DraftBackupInfo `json:"backups"`
}

// GetUserIdFromRequest resolves the access token to a user ID.
func GetUserIdFromRequest(r *http.Request) (int, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		authHeader = r.URL.Query().Get("token")
	}
	if authHeader == "" {
		return 0, fmt.Errorf("no authorization token provided")
	}
	token := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	var userId int
	err := db.DB.QueryRow("SELECT UserId FROM AccessToken WHERE Token = ?", token).Scan(&userId)
	if err != nil {
		return 0, fmt.Errorf("invalid token")
	}
	return userId, nil
}

// ListDrafts handles GET /api/drafts
func ListDrafts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utilities.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	userId, err := GetUserIdFromRequest(r)
	if err != nil {
		utilities.SendError(w, http.StatusUnauthorized, err.Error())
		return
	}

	rows, err := db.DB.Query("SELECT Id, Title, Markdown, UpdatedAt, UserId FROM Draft WHERE UserId = ? ORDER BY UpdatedAt DESC", userId)
	if err != nil {
		utilities.SendError(w, http.StatusInternalServerError, "Error querying drafts")
		return
	}
	defer rows.Close()

	drafts := []Draft{}
	for rows.Next() {
		var d Draft
		var updatedAtStr string
		err := rows.Scan(&d.Id, &d.Title, &d.Markdown, &updatedAtStr, &d.UserId)
		if err != nil {
			utilities.SendError(w, http.StatusInternalServerError, "Error scanning draft")
			return
		}

		t, err := time.Parse("2006-01-02 15:04:05", updatedAtStr)
		if err == nil {
			d.UpdatedAt = t
		} else {
			t, err = time.Parse(time.RFC3339, updatedAtStr)
			if err == nil {
				d.UpdatedAt = t
			}
		}
		drafts = append(drafts, d)
	}

	utilities.SendJSON(w, &drafts)
}

// GetDraft handles GET /api/drafts/{id}
func GetDraft(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utilities.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	userId, err := GetUserIdFromRequest(r)
	if err != nil {
		utilities.SendError(w, http.StatusUnauthorized, err.Error())
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) >= 4 {
			idStr = parts[3]
		}
	}
	draftId, err := strconv.Atoi(idStr)
	if err != nil {
		utilities.SendError(w, http.StatusBadRequest, "Invalid draft ID")
		return
	}

	var d Draft
	var updatedAtStr string
	err = db.DB.QueryRow("SELECT Id, Title, Markdown, UpdatedAt, UserId FROM Draft WHERE Id = ? AND UserId = ?", draftId, userId).Scan(&d.Id, &d.Title, &d.Markdown, &updatedAtStr, &d.UserId)
	if err == sql.ErrNoRows {
		utilities.SendError(w, http.StatusNotFound, "Draft not found")
		return
	} else if err != nil {
		utilities.SendError(w, http.StatusInternalServerError, "Error loading draft")
		return
	}

	t, err := time.Parse("2006-01-02 15:04:05", updatedAtStr)
	if err == nil {
		d.UpdatedAt = t
	} else {
		t, err = time.Parse(time.RFC3339, updatedAtStr)
		if err == nil {
			d.UpdatedAt = t
		}
	}

	// Fetch backups list
	backups := []DraftBackupInfo{}
	rows, err := db.DB.Query("SELECT Id, DraftId, Title, SavedAt FROM DraftBackup WHERE DraftId = ? ORDER BY SavedAt DESC", draftId)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var b DraftBackupInfo
			var savedAtStr string
			if err := rows.Scan(&b.Id, &b.DraftId, &b.Title, &savedAtStr); err == nil {
				savedAt, err := time.Parse("2006-01-02 15:04:05", savedAtStr)
				if err == nil {
					b.SavedAt = savedAt
				} else {
					savedAt, err = time.Parse(time.RFC3339, savedAtStr)
					if err == nil {
						b.SavedAt = savedAt
					}
				}
				backups = append(backups, b)
			}
		}
	}

	detail := DraftDetail{
		Draft:   d,
		Backups: backups,
	}
	utilities.SendJSON(w, &detail)
}

// GetDraftBackup handles GET /api/drafts/backups/{backup_id}
func GetDraftBackup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utilities.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	userId, err := GetUserIdFromRequest(r)
	if err != nil {
		utilities.SendError(w, http.StatusUnauthorized, err.Error())
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) >= 5 {
			idStr = parts[4]
		}
	}
	backupId, err := strconv.Atoi(idStr)
	if err != nil {
		utilities.SendError(w, http.StatusBadRequest, "Invalid backup ID")
		return
	}

	var d Draft
	var savedAtStr string
	err = db.DB.QueryRow(`
		SELECT db.Id, db.DraftId, db.Title, db.Markdown, db.SavedAt, d.UserId
		FROM DraftBackup db
		JOIN Draft d ON db.DraftId = d.Id
		WHERE db.Id = ? AND d.UserId = ?
	`, backupId, userId).Scan(&d.Id, &d.UserId, &d.Title, &d.Markdown, &savedAtStr, &d.UserId)

	if err == sql.ErrNoRows {
		utilities.SendError(w, http.StatusNotFound, "Backup not found")
		return
	} else if err != nil {
		utilities.SendError(w, http.StatusInternalServerError, "Error loading backup")
		return
	}

	t, err := time.Parse("2006-01-02 15:04:05", savedAtStr)
	if err == nil {
		d.UpdatedAt = t
	} else {
		t, err = time.Parse(time.RFC3339, savedAtStr)
		if err == nil {
			d.UpdatedAt = t
		}
	}

	utilities.SendJSON(w, &d)
}

// SaveDraft handles POST /api/drafts
func SaveDraft(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utilities.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	userId, err := GetUserIdFromRequest(r)
	if err != nil {
		utilities.SendError(w, http.StatusUnauthorized, err.Error())
		return
	}

	var d Draft
	success, parsed := utilities.DecodeJSON(w, r, &d)
	if !success {
		return
	}
	d = parsed

	tx, err := db.DB.Begin()
	if err != nil {
		utilities.SendError(w, http.StatusInternalServerError, "Database error: unable to start transaction")
		return
	}
	defer tx.Rollback()

	var existingId int
	isUpdate := false
	if d.Id > 0 {
		err = tx.QueryRow("SELECT Id FROM Draft WHERE Id = ? AND UserId = ?", d.Id, userId).Scan(&existingId)
		if err == nil {
			isUpdate = true
		}
	}

	now := time.Now().Format("2006-01-02 15:04:05")

	if isUpdate {
		// 1. Copy the old version to DraftBackup
		_, err = tx.Exec(`
			INSERT INTO DraftBackup (DraftId, Title, Markdown, SavedAt)
			SELECT Id, Title, Markdown, UpdatedAt
			FROM Draft WHERE Id = ?
		`, d.Id)
		if err != nil {
			utilities.SendError(w, http.StatusInternalServerError, fmt.Sprintf("Error creating version backup: %v", err))
			return
		}

		// 2. Update latest draft details
		_, err = tx.Exec(`
			UPDATE Draft
			SET Title = ?, Markdown = ?, UpdatedAt = ?
			WHERE Id = ? AND UserId = ?
		`, d.Title, d.Markdown, now, d.Id, userId)
		if err != nil {
			utilities.SendError(w, http.StatusInternalServerError, "Error updating draft")
			return
		}
	} else {
		// Insert new draft
		res, err := tx.Exec(`
			INSERT INTO Draft (Title, Markdown, UpdatedAt, UserId)
			VALUES (?, ?, ?, ?)
		`, d.Title, d.Markdown, now, userId)
		if err != nil {
			utilities.SendError(w, http.StatusInternalServerError, "Error inserting draft")
			return
		}
		lastId, err := res.LastInsertId()
		if err != nil {
			utilities.SendError(w, http.StatusInternalServerError, "Error retrieving new draft ID")
			return
		}
		d.Id = int(lastId)
	}

	err = tx.Commit()
	if err != nil {
		utilities.SendError(w, http.StatusInternalServerError, "Error committing transaction")
		return
	}

	d.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", now)
	d.UserId = userId
	utilities.SendJSON(w, &d)
}
