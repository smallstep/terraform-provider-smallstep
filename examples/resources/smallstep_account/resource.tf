
resource "smallstep_account" "wifi" {
  name = "WiFi Hosted Radius"
  wifi = {
    autojoin                 = true
    external_radius_server   = false
    hidden                   = true
    network_access_server_ip = "0.0.0.1"
    ssid                     = "corpnet"
  }
}

resource "smallstep_account" "ethernet" {
  name = "Ethernet External RADIUS"
  ethernet = {
    autojoin               = false
    external_radius_server = true
    ca_chain               = file("radius-ca.pem")
  }
}

resource "smallstep_account" "vpn" {
  name = "VPN"
  vpn = {
    autojoin        = false
    connection_type = "IKEv2"
    remote_address  = "ike.example.com"
    ike = {
      ca_chain  = file("ike-ca.pem")
      eap       = true
      remote_id = "foo"
    }
  }
}

resource "smallstep_account" "browser" {
  name    = "Browser Certificate"
  browser = {}
}

resource "smallstep_account" "generic" {
  name = "Generic Client Certificate"
  certificate = {
    duration  = "168h"
    crt_file  = "db.crt"
    key_file  = "db.key"
    root_file = "ca.crt"
    uid       = 1001
    gid       = 999
    mode      = 256
    x509 = {
      common_name = {
        static          = "example.com"
        device_metadata = "host"
      }
      sans = {
        static          = ["user@example.com"]
        device_metadata = ["sans"]
      }
      organization = {
        static = ["ops"]
      }
      organizational_unit = {
        static = ["dev"]
      }
      locality = {
        static = ["DC"]
      }
      postal_code = {
        static = ["20252"]
      }
      country = {
        device_metadata = ["country"]
      }
      street_address = {
        device_metadata = ["street"]
      }
      province = {
        device_metadata = ["province"]
      }
    }
  }
  reload = {
    method   = "SIGNAL"
    pid_file = "x.pid"
    signal   = 1
  }
  key = {
    format     = "DER"
    type       = "ECDSA_P256"
    protection = "NONE"
  }
  policy = {
    assurance = ["high"]
    os        = ["Windows", "macOS"]
    ownership = ["company"]
    source    = ["Jamf", "Intune"]
    tags      = ["mdm"]
  }
}

resource "smallstep_account" "ssh" {
  name = "SSH User Certificate"
  certificate = {
    ssh = {
      key_id = {
        device_metadata = "key"
      }
      principals = {
        static          = ["eng"]
        device_metadata = ["role"]
      }
    }
  }
}
