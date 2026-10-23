// Package oauth implements Qdrant Cloud CLI browser/device login.
//
// The Auth0 issuer is not a separate config field. It is taken from
// QDRANT_CLOUD_ENDPOINT (grpc.<cluster>.qdrant.io) by:
//  1. mapping the gRPC host to the HTTPS API host (grpc. → api.)
//  2. fetching unauthenticated RFC 9728 protected-resource metadata
//  3. falling back to login.<cluster>.qdrant.io when the gateway does not
//     advertise OAuth yet
package oauth
