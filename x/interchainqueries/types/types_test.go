package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransactionFilterValidation(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		// several conditions
		assert.NoError(t, ValidateTransactionsFilter(`[{"field":"transfer.recipient","op":"Eq","value":"neutron1mjk79fjjgpplak5wq838w0yd982gzkyf8fxu8u"},{"field":"tx.height","op":"Gte","value":100}]`))
		// all supported operations with a whole operand
		assert.NoError(t, ValidateTransactionsFilter(`[{"field":"tx.height","op":"Eq","value":1000}]`))
		assert.NoError(t, ValidateTransactionsFilter(`[{"field":"tx.height","op":"Gt","value":1000}]`))
		assert.NoError(t, ValidateTransactionsFilter(`[{"field":"tx.height","op":"Gte","value":1000}]`))
		assert.NoError(t, ValidateTransactionsFilter(`[{"field":"tx.height","op":"Lt","value":1000}]`))
		assert.NoError(t, ValidateTransactionsFilter(`[{"field":"tx.height","op":"Lte","value":1000}]`))
	})
	t.Run("Invalid", func(t *testing.T) {
		// invalid json
		assert.ErrorContains(t, ValidateTransactionsFilter(`[{"field":"transfer.recipient","op":"Eq","value":`), "unexpected end of JSON input")
		// empty operation
		assert.ErrorContains(t, ValidateTransactionsFilter(`[{"field":"transfer.recipient","op":"","value":"neutron1mjk79fjjgpplak5wq838w0yd982gzkyf8fxu8u"}]`), "op '' is expected to be one of: eq, gt, gte, lt, lte")
		// empty field
		assert.ErrorContains(t, ValidateTransactionsFilter(`[{"field":"","op":"Eq","value":"neutron1mjk79fjjgpplak5wq838w0yd982gzkyf8fxu8u"}]`), "field couldn't be empty")
		// field with forbidden symbols
		assert.ErrorContains(t, ValidateTransactionsFilter(`[{"field":"transfer.\t","op":"Eq","value":"neutron1mjk79fjjgpplak5wq838w0yd982gzkyf8fxu8u"}]`), "special symbols [\t \n \r \\ ( ) \" ' = > <] are not allowed")
		assert.ErrorContains(t, ValidateTransactionsFilter(`[{"field":"transfer.\n","op":"Eq","value":"neutron1mjk79fjjgpplak5wq838w0yd982gzkyf8fxu8u"}]`), "special symbols [\t \n \r \\ ( ) \" ' = > <] are not allowed")
		assert.ErrorContains(t, ValidateTransactionsFilter(`[{"field":"transfer.\r","op":"Eq","value":"neutron1mjk79fjjgpplak5wq838w0yd982gzkyf8fxu8u"}]`), "special symbols [\t \n \r \\ ( ) \" ' = > <] are not allowed")
		assert.ErrorContains(t, ValidateTransactionsFilter(`[{"field":"transfer.\\","op":"Eq","value":"neutron1mjk79fjjgpplak5wq838w0yd982gzkyf8fxu8u"}]`), "special symbols [\t \n \r \\ ( ) \" ' = > <] are not allowed")
		assert.ErrorContains(t, ValidateTransactionsFilter(`[{"field":"transfer.(","op":"Eq","value":"neutron1mjk79fjjgpplak5wq838w0yd982gzkyf8fxu8u"}]`), "special symbols [\t \n \r \\ ( ) \" ' = > <] are not allowed")
		assert.ErrorContains(t, ValidateTransactionsFilter(`[{"field":"transfer.)","op":"Eq","value":"neutron1mjk79fjjgpplak5wq838w0yd982gzkyf8fxu8u"}]`), "special symbols [\t \n \r \\ ( ) \" ' = > <] are not allowed")
		assert.ErrorContains(t, ValidateTransactionsFilter(`[{"field":"transfer.\"","op":"Eq","value":"neutron1mjk79fjjgpplak5wq838w0yd982gzkyf8fxu8u"}]`), "special symbols [\t \n \r \\ ( ) \" ' = > <] are not allowed")
		assert.ErrorContains(t, ValidateTransactionsFilter(`[{"field":"transfer.'","op":"Eq","value":"neutron1mjk79fjjgpplak5wq838w0yd982gzkyf8fxu8u"}]`), "special symbols [\t \n \r \\ ( ) \" ' = > <] are not allowed")
		assert.ErrorContains(t, ValidateTransactionsFilter(`[{"field":"transfer.=","op":"Eq","value":"neutron1mjk79fjjgpplak5wq838w0yd982gzkyf8fxu8u"}]`), "special symbols [\t \n \r \\ ( ) \" ' = > <] are not allowed")
		assert.ErrorContains(t, ValidateTransactionsFilter(`[{"field":"transfer.>","op":"Eq","value":"neutron1mjk79fjjgpplak5wq838w0yd982gzkyf8fxu8u"}]`), "special symbols [\t \n \r \\ ( ) \" ' = > <] are not allowed")
		assert.ErrorContains(t, ValidateTransactionsFilter(`[{"field":"transfer.<","op":"Eq","value":"neutron1mjk79fjjgpplak5wq838w0yd982gzkyf8fxu8u"}]`), "special symbols [\t \n \r \\ ( ) \" ' = > <] are not allowed")
		// decimal number
		assert.ErrorContains(t, ValidateTransactionsFilter(`[{"field":"tx.height","op":"Gte","value":15.5}]`), "can't be a decimal number")
	})
}
