package party

import (
	"fmt"
	"net/http"
	"partymanager/server/api"
	"partymanager/server/auth"
	"partymanager/server/models"
	"strconv"
	"time"

	"partymanager/server/api/guest"

	"github.com/go-chi/chi/v5"
)

func (ur *PartyRoutes) GetAllParty(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetToken(&r.Header, ur.Auth.TokenAuth)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return
	}

	id, _ := token.Get("id")
	idInt, _ := strconv.Atoi(fmt.Sprintf("%v", id)) // Convert id to integer

	email, _ := token.Get("email")

	var parties []models.Party

	ur.DB.Model(&models.Party{}).Joins("LEFT JOIN guests on guests.party_id = parties.id").Not("deleted_at IS NOT NULL").Where("host_id = ?", idInt).Or("guests.email = ?", email).Group("parties.id").Find(&parties)

	var response = []PartyResponse{}

	for _, party := range parties {
		response = append(response, PartyResponse{ID: party.ID, Name: party.Name, Description: party.Description, Location: party.Location, Date: party.Date.String(), HostID: party.HostID})
	}

	fmt.Println(response)

	api.EncodeBody(w, response)
}

func (ur *PartyRoutes) CreateParty(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetToken(&r.Header, ur.Auth.TokenAuth)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return
	}

	id, _ := token.Get("id")
	idInt, _ := strconv.Atoi(fmt.Sprintf("%v", id)) // Convert id to integer

	var body CreatePartyRequest
	err = api.DecodeBody(r, &body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	// date, err := time.Parse("2006-01-02T15:04", body.Date)
	date, err := time.Parse("2006-01-02 15:04:05 -0700", body.Date)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	party := models.Party{
		Name:        body.Name,
		Description: body.Description,
		Location:    body.Location,
		Date:        date,
		HostID:      uint(idInt),
	}

	ur.DB.Create(&party)

	api.EncodeBody(w, NewPartyResponse{ID: party.ID})
}

func (ur *PartyRoutes) GetParty(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetToken(&r.Header, ur.Auth.TokenAuth)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return
	}

	useremail, _ := token.Get("email")
	userid, _ := token.Get("id")
	useridInt, _ := strconv.Atoi(fmt.Sprintf("%v", userid))

	id := chi.URLParam(r, "id")

	var party models.Party

	ur.DB.Model(&models.Party{}).Joins("LEFT JOIN guests on guests.party_id = parties.id").Where("(host_id = ? OR guests.email = ?) AND parties.id = ?", useridInt, useremail, id).Group("parties.id").First(&party)

	if party.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Party not found"))
		return
	}

	var guests []models.Guest
	ur.DB.Where("party_id = ?", id).Order("id asc").Find(&guests)

	var guestList []guest.GuestResponse = []guest.GuestResponse{}
	for _, guestItem := range guests {
		guestList = append(guestList, guest.GuestResponse{
			ID:       guestItem.ID,
			Username: guestItem.Username,
			Email:    guestItem.Email,
			Present:  guestItem.Present,
		})
	}

	api.EncodeBody(w, PartyResponse{
		ID:          party.ID,
		Name:        party.Name,
		Description: party.Description,
		Location:    party.Location,
		Date:        party.Date.String(),
		HostID:      party.HostID,
		Guests:      guestList,
	})
}

func (ur *PartyRoutes) UpdateParty(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetToken(&r.Header, ur.Auth.TokenAuth)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return
	}

	userid, _ := token.Get("id")
	idInt, _ := strconv.Atoi(fmt.Sprintf("%v", userid)) // Convert id to integer

	id := chi.URLParam(r, "id")
	partyId, _ := strconv.Atoi(id)

	var party models.Party
	ur.DB.Where(map[string]interface{}{"host_id": idInt, "id": partyId}).First(&party)

	if party.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Party not found"))
		return
	}

	var body UpdatePartyRequest
	api.DecodeBody(r, &body)

	// date, err := time.Parse("2006-01-02T15:04", body.Date)
	date, err := time.Parse("2006-01-02 15:04:05 -0700", body.Date)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if body.Name != "" {
		party.Name = body.Name
	}
	if body.Description != "" {
		party.Description = body.Description
	}
	if body.Location != "" {
		party.Location = body.Location
	}
	if body.Date != "" {
		party.Date = date
	}

	ur.DB.Save(&party)

	api.EncodeBody(w, PartyResponse{ID: party.ID, Name: party.Name, Description: party.Description, Location: party.Location, Date: party.Date.String(), HostID: party.HostID})
}

func (ur *PartyRoutes) DeleteParty(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetToken(&r.Header, ur.Auth.TokenAuth)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return
	}

	userid, _ := token.Get("id")
	idInt, _ := strconv.Atoi(fmt.Sprintf("%v", userid)) // Convert id to integer

	id := chi.URLParam(r, "id")
	partyId, _ := strconv.Atoi(id)

	var party models.Party
	party.ID = uint(partyId)
	party.HostID = uint(idInt)

	ur.DB.First(&party)

	if party.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Party not found"))
		return
	}

	ur.DB.Model(&models.Party{}).Where("id  = ?", partyId).Update("deleted_at", time.Now())

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Party deleted"))
}

func (ur *PartyRoutes) GetSharedParty(w http.ResponseWriter, r *http.Request) {
	link := chi.URLParam(r, "link")

	var currentGuest models.Guest
	ur.DB.Where("link_token = ?", link).First(&currentGuest)

	if currentGuest.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Guest not found"))
		return
	}

	var party models.Party
	ur.DB.Where("id = ?", currentGuest.PartyID).First(&party)

	if party.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Party not found"))
		return
	}

	var guestsList []guest.GuestResponse = []guest.GuestResponse{}
	guestsList = append(guestsList, guest.GuestResponse{ID: currentGuest.ID, Username: currentGuest.Username, Email: currentGuest.Email, Present: currentGuest.Present})

	api.EncodeBody(w, PartyResponse{
		ID:          party.ID,
		Name:        party.Name,
		Description: party.Description,
		Location:    party.Location,
		Date:        party.Date.String(),
		HostID:      party.HostID,
		Guests:      guestsList,
	})
}
