package ovsdb

import (
	"fmt"
	"math/rand"

	"github.com/ebay/libovsdb"
	"github.com/fperf/fperf"
)

const (
	OP_INSERT string = "insert"
	OP_UPDATE string = "update"
	OP_DELETE string = "delete"
	OP_SELECT string = "select"
)

const (
	DBNAME string = "OVN_Southbound"
	TABLE  string = "DHCP_Options"
)

func init() {
	fperf.Register("ovsdb", New, "ovsdb benchmark")
}

// Op is the operation type
type Op string

// Operations
const (
	Put    Op = "put"
	Get    Op = "get"
	Range  Op = "range"
	Delete Op = "delete"
)

type client struct {
	ovsdb *libovsdb.OvsdbClient
	uuids []libovsdb.UUID
	op    Op
}

// New creates a fperf client
func New(fs *fperf.FlagSet) fperf.Client {
	var keySize int
	var op Op
	fs.IntVar(&keySize, "key-size", 4, "length of the random key")
	fs.Parse()
	args := fs.Args()
	if len(args) == 0 {
		op = Put
	} else {
		op = Op(args[0])
	}
	return &client{
		op:    op,
	}
}

// Dial to ovsdb
func (c *client) Dial(addr string) error {
	cli, err := libovsdb.Connect(addr, nil)
	if err != nil {
		return fmt.Errorf("Dial error: %s", err)
	}
	c.ovsdb = cli
	return initUUID(c)
}

// Request ovsdb
func (c *client) Request() error {
	switch c.op {
	case Put:
		return doPut(c)
	case Get:
		return doGet(c)
	case Range:
		return doRange(c)
	case Delete:
		return doDelete(c)
	}
	return fmt.Errorf("unknown op %s", c.op)
}

func doPut(c *client) error {
	row := map[string]interface{}{
		"name": "name",
		"code": 210,
		"type": "str",
	}
	updateOp := libovsdb.Operation{
		Table: TABLE,
		Op:    OP_INSERT,
		Row:   row,
	}
	operations := []libovsdb.Operation{updateOp}
	reply, err := c.ovsdb.Transact(DBNAME, operations...)
	// fmt.Printf("reply=%+v\n", reply)
	// fmt.Printf("err=%+v\n", err)
	return isTransactError(reply, err, operations)
}

func initUUID(c *client) error {
	c.uuids = []libovsdb.UUID{}

	selectOp := libovsdb.Operation{
		Op:    OP_SELECT,
		Table: TABLE,
	}
	operations := []libovsdb.Operation{selectOp}
	reply, err := c.ovsdb.Transact(DBNAME, operations...)
	// fmt.Printf("reply=%+v\n", reply)
	// fmt.Printf("err=%+v\n", err)
	if err = isTransactError(reply, err, operations); err != nil {
		return err
	}
	if len(reply) > 0 {
		for _, o := range reply {
			for _, r := range o.Rows {
				c.uuids = append(c.uuids, r["_uuid"].(libovsdb.UUID))
			}
		}
	}
	return nil
}

func getUUID(c *client) (libovsdb.UUID, error) {
	if len(c.uuids) == 0 {
		return libovsdb.UUID{}, nil
	}
	// fmt.Printf("c.uuids=%+v\n", c.uuids)
	return c.uuids[rand.Intn(len(c.uuids))], nil
}

func doGet(c *client) error {
	uuid, err := getUUID(c)
	if err != nil {
		return err
	}
	condition := libovsdb.NewCondition("_uuid", "==", uuid)
	selectOp := libovsdb.Operation{
		Op:    OP_SELECT,
		Table: TABLE,
		Where: []interface{}{condition},
	}
	operations := []libovsdb.Operation{selectOp}
	reply, err := c.ovsdb.Transact(DBNAME, operations...)
	// fmt.Printf("reply=%+v\n", reply)
	// fmt.Printf("err=%+v\n", err)
	return isTransactError(reply, err, operations)
}

func doDelete(c *client) error {
	return nil // TODO: missing
}

func doRange(c *client) error {
	return nil // TODO: missing
}

func isTransactError(reply []libovsdb.OperationResult, err error, operations ...[]libovsdb.Operation) error {
	if err != nil {
		return err
	}
	errStr := ""
	if len(reply) < len(operations) {
		errStr = fmt.Sprintf("Number of Replies should be atleast equal to number of Operations\n")
	}
	fail := false
	if len(reply) > 0 {
		for i, o := range reply {
			if o.Error != "" && i < len(operations) {
				errStr += fmt.Sprintf("Transaction Failed due to an error : %v details: %v in: %v\n", o.Error, o.Details, operations[i])
				fail = true
			} else if o.Error != "" {
				errStr += fmt.Sprintf("Transaction Failed due to an error : %v\n", o.Error)
				fail = true
			}
		}
	} else {
		fail = true
	}
	if fail {
		return fmt.Errorf(errStr)
	}
	return nil
}
