//go:build !custom || inputs || inputs.asterisk

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/asterisk" // register plugin
