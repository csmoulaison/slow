package main

import (
	"io/ioutil"
	"os"
	"strings"
	"encoding/csv"
	"strconv"
	"sort"
)

const UserDir = "../../data/users/"
const UserFileExt = ".ur"

type User struct {
	Handle string
	Password string
	Rolodex []string
	Email string
	NotifyByEmail bool
	MailboxCache []int
	SentCache []int
}

func (u *User) save() error {
	fname := UserDir + u.Handle + UserFileExt
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()

	sort.Strings(u.Rolodex)
	sort.Slice(u.MailboxCache, func(i, j int) bool {
        return u.MailboxCache[i] > u.MailboxCache[j]
    })
	sort.Slice(u.SentCache, func(i, j int) bool {
        return u.SentCache[i] > u.SentCache[j]
    })

	writer := csv.NewWriter(f)
	defer writer.Flush()

	// See CSV schema for user data in loadUser()
	// Conversion from bool to string for NotifyByEmail field
	notifyString := "0"
	if u.NotifyByEmail {
		notifyString = "1"
	}
	// D = Detail
	userDetails := []string{
		"D",
		u.Handle, 
		u.Password, 
		u.Email, 
		notifyString}
	writer.Write(userDetails)

	// C = Contact
	for _, handle := range u.Rolodex {
		contact := []string{
			"C",
			handle}
		writer.Write(contact)
	}

	// R = Mailbox (received) letter cached id
	for _, id := range u.MailboxCache {
		mailboxCache := []string{
			"R",
			strconv.Itoa(id)}
		writer.Write(mailboxCache)
	}

	// S = Sent letter cached id
	for _, id := range u.SentCache {
		sentCache := []string{
			"S",
			strconv.Itoa(id)}
		writer.Write(sentCache)
	}

	return nil
}

func loadUser(handle string) (User, error) {
	fname := UserDir + handle + UserFileExt
	f, err := os.Open(fname)
	if err != nil {
		return User{}, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1
	data, err := reader.ReadAll()
	if err != nil {
		return User{}, err
	}

	// There are multiple types of line, determined by the indicator at row[0].
	// 1. D,HANDLE,PASSWORD,DISPLAYNAME,EMAIL,NOTIFYBYEMAIL("0"||"1") - Detail
	// 2. C,HANDLE - Contact
	// 2. R,LETTERID - Mailbox Cache (received letters)
	// 2. S,LETTERID - Sent Cache (sent letters)
	u := User{}
	for _, row := range data {
		switch row[0] {
		case "D": // Detail
			u.Handle      = row[1]
			u.Password    = row[2]
			u.Email       = row[3]
			// Conversion from string to bool for NotifyByEmail
			if row[4] == "0" {
				u.NotifyByEmail = false
			} else {
				u.NotifyByEmail = true
			}
		case "C": // Contact
			u.Rolodex = append(u.Rolodex, row[1])
		case "R": // Mailbox (received) letters cached id
			id := 0
			id, err = strconv.Atoi(row[1])
			if err != nil {
				return User{}, err
			}
			u.MailboxCache = append(u.MailboxCache, id)
		case "S": // Sent letters cached id
			id := 0
			id, err = strconv.Atoi(row[1])
			if err != nil {
				return User{}, err
			}
			u.SentCache = append(u.SentCache, id)
		}
	}
	return u, nil
}

func allUsers() ([]User, error) {
	files, err := ioutil.ReadDir(UserDir)
	if err != nil {
		return nil, err
	}

	users := []User{}
	for _, f := range files {
		fname := f.Name()
		u, err := loadUser(fname[:strings.Index(fname, UserFileExt)])
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}
