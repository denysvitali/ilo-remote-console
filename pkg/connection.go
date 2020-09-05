package ilo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

type Connection struct {
	rcInfo *RcInfo
	client *http.Client
	socket *net.Conn

	hostname string
	username string
	password string

	sessionKey string
}

func NewCustom(c *http.Client) Connection {
	return Connection{client: c}
}

func New() Connection {
	return Connection{client: http.DefaultClient}
}

type LoginRequest struct {
	Method string `json:"method"`
	Username string `json:"user_login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	SessionKey string `json:"session_key"`
	UserName string `json:"user_name"`
	UserAccount string `json:"user_account"`
	UserDN string `json:"user_dn"`
	UserType string `json:"user_type"`
	UserIp string `json:"user_ip"`
	UserExpires IloTimestamp `json:"user_expires"`
	LoginPriv int `json:"login_priv"`
	RemoteConsPriv int `json:"remote_cons_priv"`
	VirtualMediaPriv int `json:"virtual_media_priv"`
	ResetPriv int `json:"reset_priv"`
	ConfigPriv int `json:"config_priv"`
	UserPriv int `json:"user_priv"`
}

type IloTimestamp struct {
	time time.Time
}

func (i IloTimestamp) UnmarshalJSON(data []byte) error {
	time, err := time.Parse("\"Mon Jan 02 15:04:05 2006\"", string(data))
	if err != nil {
		return err
	}

	i.time = time

	return nil
}

func (c *Connection) login() error{
	url := fmt.Sprintf("https://%s/json/login_session", c.hostname)
	loginRequest := LoginRequest{
		Method:   "login",
		Username: c.username,
		Password: c.password,
	}

	loginRequestBytes, err := json.Marshal(&loginRequest)

	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(loginRequestBytes))
	if err != nil {
		return err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var loginResponse LoginResponse

	err = json.Unmarshal(bodyBytes, &loginResponse)
	if err != nil {
		return err
	}

	c.sessionKey = loginResponse.SessionKey
	return nil
}

func (c *Connection) Connect(hostname string, username string, password string) error {
	c.hostname = hostname
	c.username = username
	c.password = password
	err := c.login()
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://%s/json/rc_info", hostname)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.AddCookie(&http.Cookie{Name: "sessionKey", Value: c.sessionKey})

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&c.rcInfo)

	if err != nil {
		return err
	}

	return nil
}

func (c* Connection) GetScreenImage() (image.Image, error) {
	if c.socket == nil {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", c.hostname, c.rcInfo.RcPort))
		if err != nil {
			return nil, err
		}

		c.socket = &conn
	}

	socket := *c.socket
	byteArr := make([]byte, 5)
	read, err := socket.Read(byteArr)
	if err != nil {
		return nil, err
	}

	if read != 1 {
		return nil, errors.New(fmt.Sprintf("unexpected number of bytes read: expected 1 but got %d", read))
	}

	if byteArr[0] != 80 {
		return nil, errors.New("invalid hello")
	}

	// request remote conn

	success, err := requestRemoteConnection(c)

	if err != nil {
		return nil, err
	}

	if !success {
		return nil, errors.New("unsuccessful remote connection request")
	}

	_, err = socket.Read(byteArr)

	if err != nil {
		return nil, errors.New("unable to read bytes")
	}

	switch byteArr[0]{
	case 81:
		return nil, errors.New("access denied")
	case 82:
		fmt.Printf("authenticated")
	case 83:
		fallthrough
	case 89:
		fmt.Printf("authenticated, busy")
	case 87:
		return nil, errors.New("no license")
	case 88:
		return nil, errors.New("no free sessions")
	}

	return nil, nil
}

func requestRemoteConnection(c *Connection) (bool, error) {
	const separator = ' '
	if c.socket == nil {
		return false, errors.New("socket is nil")
	}

	socket := *c.socket

	var connArr = make([]byte, 2)
	connArr[0] = byte(int(separator) & 0xFF)
	connArr[1] = byte(int(separator & 0xFF00) >> 8)

	originalSessionKeyBytes := []byte(c.sessionKey)
	maskedSessionKey := []byte(c.sessionKey)
	encryptionKeyBytes := []byte(c.rcInfo.EncKey)

	if len(c.rcInfo.EncKey) != 0 {
		for i:=0; i<len(c.sessionKey); i++ {
			maskedSessionKey[i] = byte(originalSessionKeyBytes[i] ^ encryptionKeyBytes[i%len(encryptionKeyBytes)])
		}
	}

	if len(c.rcInfo.VmKey) != 0 {
		connArr[1] = byte(connArr[1] | 0x40)
	} else {
		connArr[1] = byte(connArr[1] | 0x80)
	}

	connArr = append(connArr, maskedSessionKey...)

	written, err := socket.Write(connArr)
	if err != nil {
		return false, err
	}

	if written != len(connArr){
		return false, errors.New("we haven't written all of our data")
	}

	return true, nil
}

func (c *Connection) Info() *RcInfo {
	return c.rcInfo
}
