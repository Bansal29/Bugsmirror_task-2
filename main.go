// Aryan Bansal
// 1032211329
// Bugsmirror technical paper-2
// Task-3
// Complaint application backend using GoLang
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// User struct
type User struct {
	ID          string      `json:"id"`
	SecretCode  string      `json:"secret_code"`
	Name        string      `json:"name"`
	Email       string      `json:"email"`
	Role        string      `json:"role"`
	Complaints  []Complaint `json:"complaints"`
}

// Complaint struct 
type Complaint struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Summary  string `json:"summary"`
	Severity int    `json:"severity"`
	Resolved bool   `json:"resolved"`
}

// Roles
const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

// Map for users
var users = make(map[string]User)

func main() {
	seedUsers()

	// API routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/submitComplaint", submitComplaintHandler)
	http.HandleFunc("/getAllComplaintsForUser", getAllComplaintsForUserHandler)
	http.HandleFunc("/getAllComplaintsForAdmin", getAllComplaintsForAdminHandler)
	http.HandleFunc("/viewComplaint", viewComplaintHandler)
	http.HandleFunc("/resolveComplaint", resolveComplaintHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}


//Homepage handle route
func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Complaint Portal API is running"))
	fmt.Println("API running!")
}


//login handle route
func loginHandler(w http.ResponseWriter, r *http.Request) {
	var userReq struct {
		SecretCode string `json:"secret_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&userReq); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	user, ok := users[userReq.SecretCode]
	if !ok {
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	}
	respondWithJSON(w, http.StatusOK, user)
}


//register handle route
func registerHandler(w http.ResponseWriter, r *http.Request) {
	var newUser User
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Check if user with provided secret code already exists
	if _, ok := users[newUser.SecretCode]; ok {
		respondWithError(w, http.StatusConflict, "User already exists")
		return
	}

	// Check if an admin already exists
	for _, user := range users {
		if user.Role == RoleAdmin {
			respondWithError(w, http.StatusForbidden, "Admin already exists. Cannot register as admin")
			return
		}
	}
	users[newUser.SecretCode] = newUser
	respondWithJSON(w, http.StatusCreated, newUser)
}

//Submitcomplaint route
func submitComplaintHandler(w http.ResponseWriter, r *http.Request) {
	var complaint Complaint
	if err := json.NewDecoder(r.Body).Decode(&complaint); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	secretCode := r.Header.Get("X-Secret-Code")
	if secretCode == "" {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, ok := users[secretCode]
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Add complaint to user's list of complaints
	user.Complaints = append(user.Complaints, complaint)
	users[secretCode] = user

	respondWithJSON(w, http.StatusCreated, complaint)
}


//Get all complaints for the user route
func getAllComplaintsForUserHandler(w http.ResponseWriter, r *http.Request) {
	secretCode := r.Header.Get("X-Secret-Code")
	if secretCode == "" {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, ok := users[secretCode]
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	respondWithJSON(w, http.StatusOK, user.Complaints)
}


//Get all complaints for Admin route
func getAllComplaintsForAdminHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is admin
	secretCode := r.Header.Get("X-Secret-Code")
	if secretCode == "" {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, ok := users[secretCode]
	if !ok || user.Role != RoleAdmin {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Prepare list of all complaints with user details
	allComplaints := make([]map[string]interface{}, 0)
	for _, user := range users {
		for _, complaint := range user.Complaints {
			complaintData := map[string]interface{}{
				"userID":    user.ID,
				"userName":  user.Name,
				"complaint": complaint,
			}
			allComplaints = append(allComplaints, complaintData)
		}
	}

	// Respond with all complaints
	respondWithJSON(w, http.StatusOK, allComplaints)
}



//View complaint for user route
func viewComplaintHandler(w http.ResponseWriter, r *http.Request) {
	// Find user by secret code
	secretCode := r.Header.Get("X-Secret-Code")
	if secretCode == "" {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, ok := users[secretCode]
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse complaint ID from query parameter
	complaintID := r.URL.Query().Get("id")

	// Find complaint by ID
	var foundComplaint Complaint
	for _, complaint := range user.Complaints {
		if complaint.ID == complaintID {
			foundComplaint = complaint
			break
		}
	}

	// Respond with complaint details
	if foundComplaint.ID == "" {
		respondWithError(w, http.StatusNotFound, "Complaint not found")
		return
	}
	respondWithJSON(w, http.StatusOK, foundComplaint)
}


//Resolve complaint route for admin
func resolveComplaintHandler(w http.ResponseWriter, r *http.Request) {
	
	secretCode := r.Header.Get("X-Secret-Code")
	if secretCode == "" {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, ok := users[secretCode]
	if !ok || user.Role != RoleAdmin {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	complaintID := r.URL.Query().Get("id")

	// Find the user who submitted the complaint
	var userWithComplaint User
	for _, user := range users {
		for _, complaint := range user.Complaints {
			if complaint.ID == complaintID {
				userWithComplaint = user
				break
			}
		}
	}

	// Mark complaint as resolved
	for i, complaint := range userWithComplaint.Complaints {
		if complaint.ID == complaintID {
			userWithComplaint.Complaints[i].Resolved = true
			break
		}
	}

	// Update user record
	users[userWithComplaint.SecretCode] = userWithComplaint

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Complaint resolved successfully"})
}



// Utility functions for JSON response
//chatGPT help
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}


//seeding slice with dummy values for testing API
func seedUsers() {
	users["secret1"] = User{
		ID:         "1",
		SecretCode: "secret1",
		Name:       "John Doe",
		Email:      "john@example.com",
		Role:       RoleUser, 
		Complaints: []Complaint{},
	}

	users["secret2"] = User{
		ID:         "2",
		SecretCode: "secret2",
		Name:       "Jane Smith",
		Email:      "jane@example.com",
		Role:       RoleAdmin,
		Complaints: []Complaint{},
	}
}


/*Functionalities implemented
1. Register user 
2. Login user
3. Submit complaint
4. Get complaint user specific
5. get all complaint admin
6. update complaint status admin
7. View id specific complaint for user
*/