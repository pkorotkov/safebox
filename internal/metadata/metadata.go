package metadata

import (
	"os/user"

	"github.com/pkorotkov/safebox/internal/veracrypt"
)

type Metadata map[string]interface{}

func New() Metadata {
	md := make(map[string]interface{})
	md["enough-priviledges"] = false
	return md
}

func (md Metadata) VolumeInfos() []*veracrypt.VolumeInfo {
	return md["veracrypt-volumes-infos"].([]*veracrypt.VolumeInfo)
}

func (md Metadata) SetVolumeInfos(vis []*veracrypt.VolumeInfo) {
	md["veracrypt-volumes-infos"] = vis
}

func (md Metadata) EnoughPriviledges() bool {
	return md["enough-priviledges"].(bool)
}

func (md Metadata) SetEnoughPriviledges() {
	md["enough-priviledges"] = true
}

func (md Metadata) RealUser() (*user.User, bool) {
	ru, ok := md["real-user"]
	if ok && ru != nil {
		return ru.(*user.User), true
	} else {
		return nil, false
	}
}

func (md Metadata) SetRealUser(ru *user.User) {
	md["real-user"] = ru
}
