package ddblib

import (
	"../shared"
	"../util"
)

// Client - The struct for client
type Client struct {
	IsAdmin        bool
	Address        string
	LocalDirectory string
	OtherClients   []string
}

// NotifyNewClientConnection - Other Client will notify this Client a new connection
func (c *Client) NotifyNewClientConnection(args *shared.ClientInfo, reply *shared.ClientReply) error {
	//Add other client to my list
	if !util.Contains(c.OtherClients, args.Address) {
		c.OtherClients = append(c.OtherClients, args.Address)
	}
	reply.Reply = true
	return nil
}

// NotifyOtherClients - Notify other Clients
func (c *Client) NotifyOtherClients() error {
	for _, other :=  range c.OtherClients {

	}
	return nil
}
