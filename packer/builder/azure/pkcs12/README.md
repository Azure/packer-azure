This is a fork of the from the original PKCS#12 parsing code
published in the Azure repository [go-pkcs12](https://github.com/Azure/go-pkcs12).
This fork adds serializing a x509 certificate and private key as PKCS#12 binary blob
(aka .PFX file).  Due to the specific nature of this code it was not accepted for
inclusion in the official Go crypto repository.
