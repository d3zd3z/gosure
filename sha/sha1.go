// Binding to openssl SHA1 function.

package sha

// #cgo pkg-config: libcrypto
// #include <openssl/sha.h>
import "C"
import "unsafe"

type ShaContext C.SHA_CTX

func NewSha1() *ShaContext {
	context := new(ShaContext)
	C.SHA1_Init((*C.SHA_CTX)(context))
	return context
}

// TODO: Is this pointer safe from GC?
func (ctx *ShaContext) Update(data []byte) {
	C.SHA1_Update((*C.SHA_CTX)(ctx), unsafe.Pointer(&data[0]), C.size_t(len(data)))
}

func (ctx *ShaContext) Final() []byte {
	result := make([]byte, C.SHA_DIGEST_LENGTH)
	C.SHA1_Final((*C.uchar)(&result[0]), (*C.SHA_CTX)(ctx))
	return result
}
