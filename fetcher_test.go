package main

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ens "github.com/wealdtech/go-ens/v3"
)

func TestParseZone(t *testing.T) {
	tests := []struct {
		name     string
		zone     string
		zonefile string
		err      error
	}{
		{
			name: "Empty",
			zone: "example.com",
			err:  errors.New("no SOA"),
		},
		{
			name: "NoSOA",
			zone: "example.com",
			zonefile: `$TTL    86400
$ORIGIN example.com.`,
			err: errors.New("no SOA"),
		},
		{
			name: "Good",
			zone: "jgmtest1.xyz",
			zonefile: `$TTL    86400
$ORIGIN jgmtest1.xyz.
@  1D  IN  SOA ns1x.ethdns.xyz. hostmaster.ethdns.xyz. (
                  2019123101 ; serial
                  3H ; refresh
                  15 ; retry
                  1w ; expire
                  3h ; nxdomain ttl
                 )
       IN  NS     ns1x.ethdns.xyz.
       IN  NS     ns2x.ethdns.xyz.
.      IN  A      212.47.248.33
.      IN  TXT    "v=spf1 +mx -all"
.      IN  MX     10 mail.jgmtest1.xyz
www    IN  CNAME  jgmtest1.xyz.
mail   IN  A      212.47.248.33`,
		},
		{
			name: "MultipleSOAs",
			zone: "example.com",
			zonefile: `$TTL    86400
$ORIGIN example.com.
@  1D  IN  SOA ns1.example.com. hostmaster.example.com. (2019123101 3H 15 1w 3h)
@  1D  IN  SOA ns1.example.com. hostmaster.example.com. (2019123101 3H 15 1w 3h)
.      IN  A      212.47.248.33`,
			err: errors.New("multiple SOAs"),
		},
		{
			name: "MismatchedSOAs",
			zone: "example.com",
			zonefile: `$TTL    86400
$ORIGIN bad.com.
@  1D  IN  SOA ns1.bad.com. hostmaster.bad.com. (2019123101 3H 15 1w 3h)
.      IN  A      212.47.248.33`,
			err: errors.New("mismatched SOAs"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			nodeHash, err := ens.NameHash(test.zone)
			require.Nil(t, err)
			zone, err := parseZone(test.zonefile, nodeHash)
			if test.err == nil {
				require.Nil(t, err)
				assert.Equal(t, test.zone, zone)
			} else {
				require.NotNil(t, err)
				require.Equal(t, test.err.Error(), err.Error())
			}
		})
	}

}
