// Binding to openssl SHA1 function.

package sha // import "davidb.org/x/gosure/sha"

// #cgo !darwin pkg-config: libcrypto
// #cgo darwin CFLAGS: -I/usr/local/opt/openssl/include
// #cgo darwin LDFLAGS: -L/usr/local/opt/openssl/lib -lcrypto
// #include <openssl/sha.h>
import "C"
import "unsafe"

type shaContext C.SHA_CTX

func newSha1() *shaContext {
	context := new(shaContext)
	C.SHA1_Init((*C.SHA_CTX)(context))
	return context
}

// Note that the call is GC save as long as the arguments are inline,
// and not stored in any variables.
func (ctx *shaContext) update(data []byte) {
	C.SHA1_Update((*C.SHA_CTX)(ctx), unsafe.Pointer(&data[0]), C.size_t(len(data)))
}

func (ctx *shaContext) final() []byte {
	result := make([]byte, C.SHA_DIGEST_LENGTH)
	C.SHA1_Final((*C.uchar)(&result[0]), (*C.SHA_CTX)(ctx))
	return result
}
