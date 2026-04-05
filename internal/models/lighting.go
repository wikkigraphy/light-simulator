package models

// LightType categorizes photographic light sources.
type LightType string

const (
	LightTypeSpeedlight LightType = "speedlight"
	LightTypeStrobe     LightType = "strobe"
	LightTypeContinuous LightType = "continuous"
	LightTypeLED        LightType = "led_panel"
	LightTypeRingLight  LightType = "ring_light"
	LightTypeNatural    LightType = "natural"
)

// ModifierType describes light shaping tools.
type ModifierType string

const (
	ModifierSoftbox    ModifierType = "softbox"
	ModifierOctabox    ModifierType = "octabox"
	ModifierStripbox   ModifierType = "stripbox"
	ModifierUmbrella   ModifierType = "umbrella"
	ModifierBeautyDish ModifierType = "beauty_dish"
	ModifierHoneycomb  ModifierType = "honeycomb_grid"
	ModifierSnoot      ModifierType = "snoot"
	ModifierBarnDoors  ModifierType = "barn_doors"
	ModifierDiffusion  ModifierType = "diffusion_panel"
	ModifierReflector  ModifierType = "reflector"
	ModifierParabolic  ModifierType = "parabolic"
	ModifierNone       ModifierType = "none"
)

// LightRole identifies the purpose of a light in the setup.
type LightRole string

const (
	RoleKey        LightRole = "key"
	RoleFill       LightRole = "fill"
	RoleRim        LightRole = "rim"
	RoleHair       LightRole = "hair"
	RoleBackground LightRole = "background"
	RoleAccent     LightRole = "accent"
	RoleKicker     LightRole = "kicker"
)

// Position3D represents a light's spatial coordinates relative to the subject.
// X: left(-) / right(+), Y: down(-) / up(+), Z: behind(-) / front(+).
// Distance is in meters; angle is the light's aim in degrees from subject center.
type Position3D struct {
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Z        float64 `json:"z"`
	Distance float64 `json:"distance"`
	Angle    float64 `json:"angle"`
}

// Light represents a single photographic light source with its configuration.
type Light struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Type       LightType    `json:"type"`
	Modifier   ModifierType `json:"modifier"`
	Role       LightRole    `json:"role"`
	Position   Position3D   `json:"position"`
	Power      float64      `json:"power"`       // 0-100 percentage
	ColorTemp  int          `json:"color_temp"`  // Kelvin
	CRI        int          `json:"cri"`         // Color Rendering Index 0-100
	GelColor   string       `json:"gel_color"`   // CSS color or empty
	GridDegree int          `json:"grid_degree"` // Honeycomb grid angle (10, 20, 30, 40, 60)
	Feathered  bool         `json:"feathered"`   // Whether light is feathered off-axis
	Enabled    bool         `json:"enabled"`
}

// CameraSettings models the camera configuration.
type CameraSettings struct {
	FocalLength  int     `json:"focal_length"`  // mm
	Aperture     float64 `json:"aperture"`      // f-stop
	ShutterSpeed string  `json:"shutter_speed"` // e.g. "1/200"
	ISO          int     `json:"iso"`
	WhiteBalance int     `json:"white_balance"` // Kelvin
	SensorSize   string  `json:"sensor_size"`   // "full_frame", "apsc", "medium_format"
	AngleX       float64 `json:"angle_x"`       // vertical angle
	AngleY       float64 `json:"angle_y"`       // horizontal angle
	Distance     float64 `json:"distance"`      // meters from subject
}

// ShootMode categorizes the type of photography session.
type ShootMode string

const (
	ModePortrait ShootMode = "portrait"
	ModeProduct  ShootMode = "product"
	ModeFashion  ShootMode = "fashion"
	ModeFood     ShootMode = "food"
	ModeHeadshot ShootMode = "headshot"
	ModeGroup    ShootMode = "group"
	ModeBoudoir  ShootMode = "boudoir"
	ModeSport    ShootMode = "sport"
)

// Scene represents the complete lighting and camera setup.
type Scene struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Mode     ShootMode      `json:"mode"`
	Lights   []Light        `json:"lights"`
	Camera   CameraSettings `json:"camera"`
	Backdrop string         `json:"backdrop"` // CSS color or gradient
	Ambient  float64        `json:"ambient"`  // ambient light level 0-1
	Notes    string         `json:"notes"`
}

// EquipmentItem describes a specific piece of gear used in a preset.
type EquipmentItem struct {
	Role        string `json:"role"`
	Device      string `json:"device"`
	Modifier    string `json:"modifier"`
	Power       string `json:"power"`
	Placement   string `json:"placement"`
	Recommended string `json:"recommended"`
}

// Preset is a named, reusable lighting configuration.
type Preset struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Category    string          `json:"category"`
	Description string          `json:"description"`
	Equipment   []EquipmentItem `json:"equipment"`
	Scene       Scene           `json:"scene"`
	Diagram     string          `json:"diagram"`
}
