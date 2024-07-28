// Server of ID service. ID service generates distributed id
// based on increasing integer, snowflake algorithm, ...
package id_server

// Identity is the type of distributed id number used in
// the system.
type Identity interface {
	Value() uint64
	String() string
}

// IdGenerator is anything that generates distributed id number.
type IdGenerator interface {
	Generate() (Identity, error)
}

// IdGeneratorRunner is functions that keep id generator available.
type IdGeneratorRunner func() error
