/*
Copyright 2026 Red Hat

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package users provides a central registry of service UIDs/GIDs. Go
// operators import the constants directly; non-Go consumers (e.g. container
// image builds) use the generated YAML file.
package users

// ServiceUser defines a service user for container image builds.
type ServiceUser struct {
	UID    int64    `yaml:"uid"`
	GID    int64    `yaml:"gid"`
	Home   string   `yaml:"home"`
	Groups []string `yaml:"groups,omitempty"`
}

// Service UID/GID constants.
//
// This is the single source of truth for all container image users,
// matching TCIB's _SUPPORTED_USERS in uid_gid_manage.sh.
//
// WARNING: Do not change existing values. Services that write to persistent
// storage (nova, glance, cinder, swift, galera) have files owned by these
// UIDs on PersistentVolumes. Changing a UID would require a manual chown
// migration on every PV in every deployment.
const (
	AnsibleUID         int64 = 227
	AnsibleGID         int64 = 227
	AodhUID            int64 = 42402
	AodhGID            int64 = 42402
	ApacheUID          int64 = 48
	ApacheGID          int64 = 48
	BarbicanUID        int64 = 42403
	BarbicanGID        int64 = 42403
	CeilometerUID      int64 = 42405
	CeilometerGID      int64 = 42405
	CloudAdminUID      int64 = 42401
	CloudAdminGID      int64 = 42401
	CloudkittyUID      int64 = 42406
	CloudkittyGID      int64 = 42406
	CinderUID          int64 = 42407
	CinderGID          int64 = 42407
	CyborgUID          int64 = 42485
	CyborgGID          int64 = 42485
	DesignateUID       int64 = 42411
	DesignateGID       int64 = 42411
	FrrUID             int64 = 42484
	FrrGID             int64 = 42484
	FrrVtyUID          int64 = 42483
	FrrVtyGID          int64 = 42483
	GlanceUID          int64 = 42415
	GlanceGID          int64 = 42415
	HaproxyUID         int64 = 42454
	HaproxyGID         int64 = 42454
	HeatUID            int64 = 42418
	HeatGID            int64 = 42418
	HorizonUID         int64 = 42420
	HorizonGID         int64 = 42420
	HorizontestUID     int64 = 42455
	HorizontestGID     int64 = 42455
	HugetlbfsUID       int64 = 42477
	HugetlbfsGID       int64 = 42477
	IronicUID          int64 = 42422
	IronicGID          int64 = 42422
	IronicInspectorUID int64 = 42461
	IronicInspectorGID int64 = 42461
	KeystoneUID        int64 = 42425
	KeystoneGID        int64 = 42425
	KollaUID           int64 = 42400
	KollaGID           int64 = 42400
	LibvirtUID         int64 = 42473
	LibvirtGID         int64 = 42473
	ManilaUID          int64 = 42429
	ManilaGID          int64 = 42429
	MemcachedUID       int64 = 42457
	MemcachedGID       int64 = 42457
	MysqlUID           int64 = 42434
	MysqlGID           int64 = 42434
	NamedUID           int64 = 25
	NamedGID           int64 = 25
	NeutronUID         int64 = 42435
	NeutronGID         int64 = 42435
	NfastUID           int64 = 42481
	NfastGID           int64 = 42481
	NovaUID            int64 = 42436
	NovaGID            int64 = 42436
	OctaviaUID         int64 = 42437
	OctaviaGID         int64 = 42437
	OpenvswitchUID     int64 = 42476
	OpenvswitchGID     int64 = 42476
	OvnBgpUID          int64 = 42486
	OvnBgpGID          int64 = 42486
	PlacementUID       int64 = 42482
	PlacementGID       int64 = 42482
	QemuUID            int64 = 107
	QemuGID            int64 = 107
	RabbitmqUID        int64 = 42439
	RabbitmqGID        int64 = 42439
	RallyUID           int64 = 42440
	RallyGID           int64 = 42440
	RedisUID           int64 = 42460
	RedisGID           int64 = 42460
	SwiftUID           int64 = 42445
	SwiftGID           int64 = 42445
	TempestUID         int64 = 42480
	TempestGID         int64 = 42480
	TobikoUID          int64 = 42495
	TobikoGID          int64 = 42495
	TssUID             int64 = 59
	TssGID             int64 = 59
	ValkeyUID          int64 = 42460
	ValkeyGID          int64 = 42460
	WatcherUID         int64 = 42451
	WatcherGID         int64 = 42451
)

// Registry maps service names to their user definitions. This is the
// source of truth — the YAML generator (gen_uid_gid_yaml.go) marshals
// this to zz_generated_uid_gid.yaml for container image builds.
// Read-only at runtime; do not modify.
var Registry = map[string]ServiceUser{
	"ansible":          {UID: AnsibleUID, GID: AnsibleGID, Home: "/var/lib/ansible", Groups: []string{"kolla"}},
	"aodh":             {UID: AodhUID, GID: AodhGID, Home: "/var/lib/aodh", Groups: []string{"kolla"}},
	"apache":           {UID: ApacheUID, GID: ApacheGID},
	"barbican":         {UID: BarbicanUID, GID: BarbicanGID, Home: "/var/lib/barbican", Groups: []string{"kolla", "nfast"}},
	"ceilometer":       {UID: CeilometerUID, GID: CeilometerGID, Home: "/var/lib/ceilometer", Groups: []string{"kolla"}},
	"cinder":           {UID: CinderUID, GID: CinderGID, Home: "/var/lib/cinder", Groups: []string{"kolla"}},
	"cloud-admin":      {UID: CloudAdminUID, GID: CloudAdminGID, Home: "/home/cloud-admin", Groups: []string{"kolla"}},
	"cloudkitty":       {UID: CloudkittyUID, GID: CloudkittyGID, Home: "/var/lib/cloudkitty", Groups: []string{"kolla"}},
	"cyborg":           {UID: CyborgUID, GID: CyborgGID, Home: "/var/lib/cyborg", Groups: []string{"kolla"}},
	"designate":        {UID: DesignateUID, GID: DesignateGID, Home: "/var/lib/designate", Groups: []string{"kolla"}},
	"frr":              {UID: FrrUID, GID: FrrGID, Home: "/var/run/frr", Groups: []string{"kolla", "frrvty"}},
	"frrvty":           {UID: FrrVtyUID, GID: FrrVtyGID},
	"glance":           {UID: GlanceUID, GID: GlanceGID, Home: "/var/lib/glance", Groups: []string{"kolla"}},
	"haproxy":          {UID: HaproxyUID, GID: HaproxyGID, Home: "/var/lib/haproxy", Groups: []string{"kolla"}},
	"heat":             {UID: HeatUID, GID: HeatGID, Home: "/var/lib/heat", Groups: []string{"kolla"}},
	"horizon":          {UID: HorizonUID, GID: HorizonGID, Home: "/var/lib/horizon", Groups: []string{"kolla"}},
	"horizontest":      {UID: HorizontestUID, GID: HorizontestGID, Home: "/var/lib/horizontest", Groups: []string{"kolla"}},
	"hugetlbfs":        {UID: HugetlbfsUID, GID: HugetlbfsGID},
	"ironic":           {UID: IronicUID, GID: IronicGID, Home: "/var/lib/ironic", Groups: []string{"kolla"}},
	"ironic-inspector": {UID: IronicInspectorUID, GID: IronicInspectorGID, Home: "/var/lib/ironic-inspector", Groups: []string{"kolla"}},
	"keystone":         {UID: KeystoneUID, GID: KeystoneGID, Home: "/var/lib/keystone", Groups: []string{"kolla", "apache"}},
	"kolla":            {UID: KollaUID, GID: KollaGID},
	"libvirt":          {UID: LibvirtUID, GID: LibvirtGID},
	"manila":           {UID: ManilaUID, GID: ManilaGID, Home: "/var/lib/manila", Groups: []string{"kolla"}},
	"memcached":        {UID: MemcachedUID, GID: MemcachedGID, Home: "/run/memcache", Groups: []string{"kolla"}},
	"mysql":            {UID: MysqlUID, GID: MysqlGID, Home: "/var/lib/mysql", Groups: []string{"kolla"}},
	"named":            {UID: NamedUID, GID: NamedGID, Home: "/var/named"},
	"neutron":          {UID: NeutronUID, GID: NeutronGID, Home: "/var/lib/neutron", Groups: []string{"kolla"}},
	"nfast":            {UID: NfastUID, GID: NfastGID},
	"nova":             {UID: NovaUID, GID: NovaGID, Home: "/var/lib/nova", Groups: []string{"qemu", "libvirt", "tss", "kolla"}},
	"octavia":          {UID: OctaviaUID, GID: OctaviaGID, Home: "/var/lib/octavia", Groups: []string{"kolla"}},
	"openvswitch":      {UID: OpenvswitchUID, GID: OpenvswitchGID},
	"ovn-bgp":          {UID: OvnBgpUID, GID: OvnBgpGID, Home: "/var/lib/ovn-bgp", Groups: []string{"kolla"}},
	"placement":        {UID: PlacementUID, GID: PlacementGID, Home: "/var/lib/placement", Groups: []string{"kolla"}},
	"qemu":             {UID: QemuUID, GID: QemuGID},
	"rabbitmq":         {UID: RabbitmqUID, GID: RabbitmqGID, Home: "/var/lib/rabbitmq", Groups: []string{"kolla"}},
	"rally":            {UID: RallyUID, GID: RallyGID, Home: "/var/lib/rally", Groups: []string{"kolla"}},
	"redis":            {UID: RedisUID, GID: RedisGID, Home: "/run/redis", Groups: []string{"kolla"}},
	"swift":            {UID: SwiftUID, GID: SwiftGID, Home: "/var/lib/swift", Groups: []string{"kolla"}},
	"tempest":          {UID: TempestUID, GID: TempestGID, Home: "/var/lib/tempest", Groups: []string{"kolla"}},
	"tobiko":           {UID: TobikoUID, GID: TobikoGID, Home: "/var/lib/tobiko", Groups: []string{"kolla"}},
	"tss":              {UID: TssUID, GID: TssGID},
	"valkey":           {UID: ValkeyUID, GID: ValkeyGID, Home: "/run/valkey", Groups: []string{"kolla"}},
	"watcher":          {UID: WatcherUID, GID: WatcherGID, Home: "/var/lib/watcher", Groups: []string{"kolla"}},
}
