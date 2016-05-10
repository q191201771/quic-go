package quic

import (
	"bytes"

	"github.com/lucas-clemente/quic-go/protocol"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Public Header", func() {
	Context("when parsing", func() {
		It("accepts a sample client header", func() {
			b := bytes.NewReader([]byte{0x09, 0xf6, 0x19, 0x86, 0x66, 0x9b, 0x9f, 0xfa, 0x4c, 0x51, 0x30, 0x33, 0x30, 0x01})
			publicHeader, err := ParsePublicHeader(b)
			Expect(err).ToNot(HaveOccurred())
			Expect(publicHeader.VersionFlag).To(BeTrue())
			Expect(publicHeader.ResetFlag).To(BeFalse())
			Expect(publicHeader.ConnectionID).To(Equal(protocol.ConnectionID(0x4cfa9f9b668619f6)))
			Expect(publicHeader.VersionNumber).To(Equal(protocol.VersionNumber(30)))
			Expect(publicHeader.PacketNumber).To(Equal(protocol.PacketNumber(1)))
			Expect(b.Len()).To(BeZero())
		})

		It("does not accept 0-byte connection ID", func() {
			b := bytes.NewReader([]byte{0x00, 0x01})
			_, err := ParsePublicHeader(b)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(errReceivedTruncatedConnectionID))
		})

		It("rejects 0 as a connection ID", func() {
			b := bytes.NewReader([]byte{0x09, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x51, 0x30, 0x33, 0x30, 0x01})
			_, err := ParsePublicHeader(b)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(errInvalidConnectionID))
		})

		It("accepts 1-byte packet numbers", func() {
			b := bytes.NewReader([]byte{0x08, 0xf6, 0x19, 0x86, 0x66, 0x9b, 0x9f, 0xfa, 0x4c, 0xde})
			publicHeader, err := ParsePublicHeader(b)
			Expect(err).ToNot(HaveOccurred())
			Expect(publicHeader.PacketNumber).To(Equal(protocol.PacketNumber(0xde)))
			Expect(b.Len()).To(BeZero())
		})

		It("accepts 2-byte packet numbers", func() {
			b := bytes.NewReader([]byte{0x18, 0xf6, 0x19, 0x86, 0x66, 0x9b, 0x9f, 0xfa, 0x4c, 0xde, 0xca})
			publicHeader, err := ParsePublicHeader(b)
			Expect(err).ToNot(HaveOccurred())
			Expect(publicHeader.PacketNumber).To(Equal(protocol.PacketNumber(0xcade)))
			Expect(b.Len()).To(BeZero())
		})

		It("accepts 4-byte packet numbers", func() {
			b := bytes.NewReader([]byte{0x28, 0xf6, 0x19, 0x86, 0x66, 0x9b, 0x9f, 0xfa, 0x4c, 0xad, 0xfb, 0xca, 0xde})
			publicHeader, err := ParsePublicHeader(b)
			Expect(err).ToNot(HaveOccurred())
			Expect(publicHeader.PacketNumber).To(Equal(protocol.PacketNumber(0xdecafbad)))
			Expect(b.Len()).To(BeZero())
		})

		It("accepts 6-byte packet numbers", func() {
			b := bytes.NewReader([]byte{0x38, 0xf6, 0x19, 0x86, 0x66, 0x9b, 0x9f, 0xfa, 0x4c, 0x23, 0x42, 0xad, 0xfb, 0xca, 0xde})
			publicHeader, err := ParsePublicHeader(b)
			Expect(err).ToNot(HaveOccurred())
			Expect(publicHeader.PacketNumber).To(Equal(protocol.PacketNumber(0xdecafbad4223)))
			Expect(b.Len()).To(BeZero())
		})

		PIt("rejects diversification nonces", func() {
			b := bytes.NewReader([]byte{0x0c, 0xf6, 0x19, 0x86, 0x66, 0x9b, 0x9f, 0xfa, 0x4c,
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1,
				0x01,
			})
			_, err := ParsePublicHeader(b)
			Expect(err).To(MatchError("diversification nonces should only be sent by servers"))
		})
	})

	Context("when writing", func() {
		It("writes a sample header", func() {
			b := &bytes.Buffer{}
			publicHeader := PublicHeader{
				ConnectionID:    0x4cfa9f9b668619f6,
				PacketNumber:    2,
				PacketNumberLen: protocol.PacketNumberLen6,
			}
			publicHeader.WritePublicHeader(b)
			Expect(b.Bytes()).To(Equal([]byte{0x38 | 0x04, 0xf6, 0x19, 0x86, 0x66, 0x9b, 0x9f, 0xfa, 0x4c, 2, 0, 0, 0, 0, 0}))
		})

		It("sets the Version Flag", func() {
			b := &bytes.Buffer{}
			publicHeader := PublicHeader{
				VersionFlag:     true,
				ConnectionID:    0x4cfa9f9b668619f6,
				PacketNumber:    2,
				PacketNumberLen: protocol.PacketNumberLen6,
			}
			publicHeader.WritePublicHeader(b)
			// must be the first assertion
			Expect(b.Len()).To(Equal(1 + 8)) // 1 FlagByte + 8 ConnectionID
			firstByte, _ := b.ReadByte()
			Expect(firstByte & 0x01).To(Equal(uint8(1)))
		})

		It("sets the Reset Flag", func() {
			b := &bytes.Buffer{}
			publicHeader := PublicHeader{
				ResetFlag:       true,
				ConnectionID:    0x4cfa9f9b668619f6,
				PacketNumber:    2,
				PacketNumberLen: protocol.PacketNumberLen6,
			}
			publicHeader.WritePublicHeader(b)
			// must be the first assertion
			Expect(b.Len()).To(Equal(1 + 8)) // 1 FlagByte + 8 ConnectionID
			firstByte, _ := b.ReadByte()
			Expect((firstByte & 0x02) >> 1).To(Equal(uint8(1)))
		})

		It("throws an error if both Reset Flag and Version Flag are set", func() {
			b := &bytes.Buffer{}
			publicHeader := PublicHeader{
				VersionFlag:     true,
				ResetFlag:       true,
				ConnectionID:    0x4cfa9f9b668619f6,
				PacketNumber:    2,
				PacketNumberLen: protocol.PacketNumberLen6,
			}
			err := publicHeader.WritePublicHeader(b)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(errResetAndVersionFlagSet))
		})

		It("truncates the connection ID", func() {
			b := &bytes.Buffer{}
			publicHeader := PublicHeader{
				ConnectionID:         0x4cfa9f9b668619f6,
				TruncateConnectionID: true,
				PacketNumberLen:      protocol.PacketNumberLen6,
			}
			err := publicHeader.WritePublicHeader(b)
			Expect(err).ToNot(HaveOccurred())
			firstByte, _ := b.ReadByte()
			Expect(firstByte & 0x08).To(BeZero())
			Expect(b.Bytes()).ToNot(ContainSubstring(string([]byte{0xf6, 0x19, 0x86, 0x66, 0x9b, 0x9f, 0xfa, 0x4c})))
		})

		Context("GetLength", func() {
			It("errors when calling GetLength for Version Negotiation packets", func() {
				publicHeader := PublicHeader{VersionFlag: true}
				_, err := publicHeader.GetLength()
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(errGetLengthOnlyForRegularPackets))
			})

			It("errors when calling GetLength for Public Reset packets", func() {
				publicHeader := PublicHeader{ResetFlag: true}
				_, err := publicHeader.GetLength()
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(errGetLengthOnlyForRegularPackets))
			})

			It("errors when PacketNumberLen is not set", func() {
				publicHeader := PublicHeader{
					ConnectionID: 0x4cfa9f9b668619f6,
					PacketNumber: 0xDECAFBAD,
				}
				_, err := publicHeader.GetLength()
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(errPacketNumberLenNotSet))
			})

			It("gets the length of a packet with longest packet number length and connectionID", func() {
				publicHeader := PublicHeader{
					ConnectionID:    0x4cfa9f9b668619f6,
					PacketNumber:    0xDECAFBAD,
					PacketNumberLen: protocol.PacketNumberLen6,
				}
				length, err := publicHeader.GetLength()
				Expect(err).ToNot(HaveOccurred())
				Expect(length).To(Equal(uint8(1 + 8 + 6))) // 1 byte public flag, 8 bytes connectionID, and packet number
			})

			It("gets the length of a packet with longest packet number length and truncated connectionID", func() {
				publicHeader := PublicHeader{
					ConnectionID:         0x4cfa9f9b668619f6,
					TruncateConnectionID: true,
					PacketNumber:         0xDECAFBAD,
					PacketNumberLen:      protocol.PacketNumberLen6,
				}
				length, err := publicHeader.GetLength()
				Expect(err).ToNot(HaveOccurred())
				Expect(length).To(Equal(uint8(1 + 6))) // 1 byte public flag, and packet number
			})

			It("gets the length of a packet 2 byte packet number length ", func() {
				publicHeader := PublicHeader{
					ConnectionID:    0x4cfa9f9b668619f6,
					PacketNumber:    0xDECAFBAD,
					PacketNumberLen: protocol.PacketNumberLen2,
				}
				length, err := publicHeader.GetLength()
				Expect(err).ToNot(HaveOccurred())
				Expect(length).To(Equal(uint8(1 + 8 + 2))) // 1 byte public flag, 8 byte connectionID, and packet number
			})
		})

		Context("packet number length", func() {
			It("doesn't write a header if the packet number length is not set", func() {
				b := &bytes.Buffer{}
				publicHeader := PublicHeader{
					ConnectionID: 0x4cfa9f9b668619f6,
					PacketNumber: 0xDECAFBAD,
				}
				err := publicHeader.WritePublicHeader(b)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(errPacketNumberLenNotSet))
			})

			It("writes a header with a 1-byte packet number", func() {
				b := &bytes.Buffer{}
				publicHeader := PublicHeader{
					ConnectionID:    0x4cfa9f9b668619f6,
					PacketNumber:    0xDECAFBAD,
					PacketNumberLen: protocol.PacketNumberLen1,
				}
				err := publicHeader.WritePublicHeader(b)
				Expect(err).ToNot(HaveOccurred())
				Expect(b.Bytes()).To(Equal([]byte{0x08 | 0x04, 0xf6, 0x19, 0x86, 0x66, 0x9b, 0x9f, 0xfa, 0x4c, 0xAD}))
			})

			It("writes a header with a 2-byte packet number", func() {
				b := &bytes.Buffer{}
				publicHeader := PublicHeader{
					ConnectionID:    0x4cfa9f9b668619f6,
					PacketNumber:    0xDECAFBAD,
					PacketNumberLen: protocol.PacketNumberLen2,
				}
				err := publicHeader.WritePublicHeader(b)
				Expect(err).ToNot(HaveOccurred())
				Expect(b.Bytes()).To(Equal([]byte{0x18 | 0x04, 0xf6, 0x19, 0x86, 0x66, 0x9b, 0x9f, 0xfa, 0x4c, 0xAD, 0xFB}))
			})

			It("writes a header with a 4-byte packet number", func() {
				b := &bytes.Buffer{}
				publicHeader := PublicHeader{
					ConnectionID:    0x4cfa9f9b668619f6,
					PacketNumber:    0x13DECAFBAD,
					PacketNumberLen: protocol.PacketNumberLen4,
				}
				err := publicHeader.WritePublicHeader(b)
				Expect(err).ToNot(HaveOccurred())
				Expect(b.Bytes()).To(Equal([]byte{0x28 | 0x04, 0xf6, 0x19, 0x86, 0x66, 0x9b, 0x9f, 0xfa, 0x4c, 0xAD, 0xFB, 0xCA, 0xDE}))
			})

			It("writes a header with a 6-byte packet number", func() {
				b := &bytes.Buffer{}
				publicHeader := PublicHeader{
					ConnectionID:    0x4cfa9f9b668619f6,
					PacketNumber:    0xBE1337DECAFBAD,
					PacketNumberLen: protocol.PacketNumberLen6,
				}
				err := publicHeader.WritePublicHeader(b)
				Expect(err).ToNot(HaveOccurred())
				Expect(b.Bytes()).To(Equal([]byte{0x38 | 0x04, 0xf6, 0x19, 0x86, 0x66, 0x9b, 0x9f, 0xfa, 0x4c, 0xAD, 0xFB, 0xCA, 0xDE, 0x37, 0x13}))
			})
		})
	})
})
