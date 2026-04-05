package cheatsheet

// FlashGuide contains structured flash selection guidance.
type FlashGuide struct {
	Category    string   `json:"category"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	BestFor     []string `json:"best_for"`
	PowerRange  string   `json:"power_range"`
	Recycle     string   `json:"recycle_time"`
	Tips        []string `json:"tips"`
}

// ModifierGuide describes a light modifier and its photographic applications.
type ModifierGuide struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	SizeRange  string   `json:"size_range"`
	Softness   string   `json:"softness"`
	SpillCtrl  string   `json:"spill_control"`
	BestFor    []string `json:"best_for"`
	Catchlight string   `json:"catchlight_shape"`
	ProTips    []string `json:"pro_tips"`
}

// LensGuide provides lens selection guidance for different shoot types.
type LensGuide struct {
	FocalLength string   `json:"focal_length"`
	Type        string   `json:"type"`
	BestFor     []string `json:"best_for"`
	DOFNotes    string   `json:"dof_notes"`
	Distortion  string   `json:"distortion"`
}

// FlashGuides returns professional flash selection cheatsheet data.
func FlashGuides() []FlashGuide {
	return []FlashGuide{
		{
			Category: "speedlight", Title: "Speedlight / Hotshoe Flash",
			Description: "Compact, battery-powered flash. GN 36-60. Portable and versatile.",
			BestFor:     []string{"Run-and-gun events", "On-camera bounce", "Small portable setups", "Outdoor fill flash"},
			PowerRange:  "1/128 to 1/1 (60-200Ws equivalent)",
			Recycle:     "0.5-3s depending on power",
			Tips: []string{
				"Bounce off ceilings/walls for instant soft light",
				"Use rear-curtain sync for motion blur with sharp subject",
				"HSS mode for wide apertures in daylight (f/1.4-2.8 outdoors)",
				"Stack 2-3 speedlights in a softbox for strobe-like power",
				"Use CTO gel to match tungsten ambient for natural blending",
			},
		},
		{
			Category: "monolight", Title: "Studio Monolight / Strobe",
			Description: "Self-contained studio flash. 100-1000Ws. Built-in modeling lamp.",
			BestFor:     []string{"Studio portraits", "Fashion", "Product photography", "Any controlled environment"},
			PowerRange:  "100-1000Ws",
			Recycle:     "0.1-2s",
			Tips: []string{
				"Use modeling lamp to preview light pattern before firing",
				"Set power in 1/10-stop increments for precision",
				"Pair with light meter: incident reading at subject for accurate exposure",
				"Color consistency: let flash fully recycle before firing",
				"Lower power = shorter flash duration = sharper freeze",
			},
		},
		{
			Category: "pack_head", Title: "Power Pack & Head System",
			Description: "Separate power unit and flash heads. 1200-4800Ws. Asymmetric power distribution.",
			BestFor:     []string{"High-volume commercial", "Large sets", "Multi-light precision", "Automotive / large product"},
			PowerRange:  "1200-4800Ws per pack",
			Recycle:     "0.05-1.5s",
			Tips: []string{
				"Distribute power asymmetrically: 2:1 or 3:1 between heads",
				"Use pack's modeling proportional mode for accurate preview",
				"Run at lower power for faster recycle in rapid-fire sessions",
				"Cable management critical: tape down all cables on set",
			},
		},
		{
			Category: "battery_strobe", Title: "Battery-Powered Strobe",
			Description: "Portable strobe with rechargeable battery. 200-600Ws. HSS capable.",
			BestFor:     []string{"Location portraits", "Outdoor fashion", "Overpowering sun", "Destination weddings"},
			PowerRange:  "200-600Ws",
			Recycle:     "0.01-1.5s",
			Tips: []string{
				"Full-power capacity: ~350-500 shots per charge",
				"HSS up to 1/8000s for shallow DOF in bright sun",
				"Freeze mode: t0.5 down to 1/19000s at minimum power",
				"Carry spare battery; cold weather reduces capacity 20-40%",
				"Use with grid or barn doors on location to control spill into background",
			},
		},
		{
			Category: "continuous_led", Title: "Continuous LED Panel / COB",
			Description: "Constant light output. Bi-color or RGB. 60-300W. WYSIWYG lighting.",
			BestFor:     []string{"Video + photo hybrid", "Beginners learning light", "Long exposure effects", "Product with reflections"},
			PowerRange:  "60-300W (LED equivalent)",
			Recycle:     "N/A (continuous)",
			Tips: []string{
				"WYSIWYG: what you see is what you get—no guessing",
				"Bi-color 2700-6500K range covers warm to daylight",
				"CRI 95+ essential for accurate skin tones",
				"Dimming below 10% may shift color temperature on cheaper units",
				"Ideal for managing reflections in product shots: slow adjustment",
				"Watch for flicker at certain Hz with video—use DC driver LEDs",
			},
		},
		{
			Category: "ring_light", Title: "Ring Light",
			Description: "Circular light source that wraps around the lens axis. Even, shadowless front light.",
			BestFor:     []string{"Beauty close-ups", "Macro", "Catchlight rings", "Social media content"},
			PowerRange:  "40-100W (LED) or 400Ws (flash ring)",
			Recycle:     "Varies",
			Tips: []string{
				"Creates signature circular catchlights in eyes",
				"Keep subject within 1-2m for effective wrap",
				"Combine with rear accent lights to avoid flat lighting",
				"Use as fill in combination with a key light for beauty work",
			},
		},
	}
}

// ModifierGuides returns the modifier cheatsheet data.
func ModifierGuides() []ModifierGuide {
	return []ModifierGuide{
		{
			Name: "Softbox (Rectangular)", Type: "softbox",
			SizeRange: "24×32\" to 54×72\"", Softness: "High",
			SpillCtrl:  "Moderate (add grid for more control)",
			BestFor:    []string{"Portraits", "Products", "General studio work"},
			Catchlight: "Rectangular",
			ProTips: []string{
				"Larger = softer. Double the size or halve the distance for 2-stop softer light",
				"Recessed front keeps spill off background",
				"Inner baffle adds second diffusion layer for even spread",
				"Angle slightly: 'feathering' the softbox creates smoother falloff on subject",
			},
		},
		{
			Name: "Octabox", Type: "octabox",
			SizeRange: "36\" to 7ft", Softness: "Very High",
			SpillCtrl:  "Low (wide spread, use egg-crate grid)",
			BestFor:    []string{"Beauty", "Fashion", "Full-length portraits", "Wrapping light"},
			Catchlight: "Octagonal (near-circular)",
			ProTips: []string{
				"Creates most natural-looking catchlights",
				"7ft octa simulates window light beautifully",
				"Use egg-crate grid to contain spill while keeping softness",
				"Boom overhead for clamshell base or butterfly setup",
			},
		},
		{
			Name: "Strip Box", Type: "stripbox",
			SizeRange: "9×36\" to 14×60\"", Softness: "Medium-High",
			SpillCtrl:  "Good (inherently directional)",
			BestFor:    []string{"Rim/edge light", "Product highlights", "Full-length fashion", "Hair light"},
			Catchlight: "Narrow rectangular",
			ProTips: []string{
				"Perfect for specular highlights on glossy products",
				"Two strips at 45° behind subject = classic rim setup",
				"Add grid for razor-sharp edge lighting",
				"Vertical strip beside subject creates beautiful gradation across body",
			},
		},
		{
			Name: "Beauty Dish", Type: "beauty_dish",
			SizeRange: "16\" to 28\"", Softness: "Medium",
			SpillCtrl:  "Good",
			BestFor:    []string{"Beauty / cosmetic work", "Headshots", "Fashion close-ups"},
			Catchlight: "Circular with center shadow",
			ProTips: []string{
				"White interior = softer; silver interior = more contrast and punch",
				"Add sock (diffusion cover) for softer but still directional light",
				"Keep within 2-3ft of subject for best effect",
				"Overhead position with chin-up creates classic beauty look",
			},
		},
		{
			Name: "Umbrella (Shoot-Through)", Type: "umbrella",
			SizeRange: "33\" to 65\"", Softness: "High",
			SpillCtrl:  "Poor (light spills everywhere)",
			BestFor:    []string{"Events", "Group photos", "Quick soft light", "Budget-friendly"},
			Catchlight: "Circular",
			ProTips: []string{
				"Fastest soft light to set up—seconds, not minutes",
				"Large umbrellas create beautiful wrapping light",
				"Reflective umbrella = more power efficiency, harder than shoot-through",
				"Close the umbrella partially to narrow beam spread",
			},
		},
		{
			Name: "Honeycomb Grid", Type: "honeycomb_grid",
			SizeRange: "10° to 60° (beam angle)", Softness: "Low (adds control, not softness)",
			SpillCtrl:  "Excellent",
			BestFor:    []string{"Hair light", "Accent/kicker", "Spot control", "Background spots"},
			Catchlight: "Depends on base modifier",
			ProTips: []string{
				"10° = tight spot (dramatic accent); 40° = moderate control",
				"Grid on softbox = directional soft light (best of both worlds)",
				"Essential for keeping light off background in low-key work",
				"Stack: gel + grid for colored accent with zero spill",
			},
		},
		{
			Name: "Snoot", Type: "snoot",
			SizeRange: "N/A (conical)", Softness: "Very Low (hard spotlight)",
			SpillCtrl:  "Excellent",
			BestFor:    []string{"Hair highlights", "Background spots", "Accent details", "Dramatic accents"},
			Catchlight: "Small point",
			ProTips: []string{
				"Creates hard-edged pool of light",
				"Perfect for illuminating specific details in product shots",
				"DIY: roll black cinefoil into cone for emergency snoot",
				"Combine with colored gel for artistic accent spots on background",
			},
		},
		{
			Name: "Barn Doors", Type: "barn_doors",
			SizeRange: "N/A (4-leaf)", Softness: "Low",
			SpillCtrl:  "Very Good (4-axis control)",
			BestFor:    []string{"Stage/theater", "Controlled spill", "Background shaping", "Dramatic slits"},
			Catchlight: "Rectangular (if visible)",
			ProTips: []string{
				"4 independently adjustable leaves for precise light shaping",
				"Create dramatic light slits for film-noir effects",
				"Use to flag light off parts of the scene selectively",
				"Common in video/cinema lighting for scene control",
			},
		},
		{
			Name: "Parabolic Reflector", Type: "parabolic",
			SizeRange: "28\" to 88\" (deep)", Softness: "Variable (focus-dependent)",
			SpillCtrl:  "Very Good (focusable)",
			BestFor:    []string{"Fashion editorial", "Dramatic portraits", "Commercial beauty"},
			Catchlight: "Bright ring / parabolic shape",
			ProTips: []string{
				"Move flash head within the parabolic for focus control",
				"Fully recessed = harder, more specular; forward = softer wrap",
				"True parabolics (Broncolor, Briese) vs. cheap 'deep' umbrellas differ massively",
				"Signature look: contrasty yet wrapping—favored by top fashion photographers",
			},
		},
		{
			Name: "Diffusion Panel / Scrim", Type: "diffusion_panel",
			SizeRange: "4×4ft to 12×12ft", Softness: "Very High",
			SpillCtrl:  "Low",
			BestFor:    []string{"Outdoor portraits", "Product with large even source", "Simulating window light"},
			Catchlight: "Large rectangular",
			ProTips: []string{
				"Fire strobe through panel = giant softbox without weight on stand",
				"Use outdoors to diffuse harsh sunlight",
				"Multiple layers of diffusion for progressively softer light",
				"Combine with negative fill (black flag opposite side) for dimension",
			},
		},
	}
}

// LensGuides returns lens selection cheatsheet data.
func LensGuides() []LensGuide {
	return []LensGuide{
		{FocalLength: "24mm", Type: "Wide", BestFor: []string{"Environmental portraits", "Real estate", "Group shots"}, DOFNotes: "Very deep DOF even wide open", Distortion: "Noticeable barrel distortion—keep subject centered"},
		{FocalLength: "35mm", Type: "Wide-Normal", BestFor: []string{"Street photography", "Environmental portraits", "Documentary"}, DOFNotes: "Moderate DOF at f/1.4-2", Distortion: "Minimal—versatile focal length"},
		{FocalLength: "50mm", Type: "Normal", BestFor: []string{"General portraits", "Product photography", "Walk-around"}, DOFNotes: "Natural perspective, good subject isolation at f/1.4", Distortion: "Nearly none—closest to human eye"},
		{FocalLength: "85mm", Type: "Short Telephoto", BestFor: []string{"Headshots", "Beauty", "Fashion", "Portraits"}, DOFNotes: "Beautiful bokeh, strong subject isolation", Distortion: "Flattering compression—gold standard portrait lens"},
		{FocalLength: "100-105mm", Type: "Macro/Portrait", BestFor: []string{"Macro product", "Tight headshots", "Food", "Jewelry"}, DOFNotes: "Very shallow at close focus; use f/8-11 for product", Distortion: "None—ideal for products requiring accuracy"},
		{FocalLength: "135mm", Type: "Telephoto", BestFor: []string{"Fashion runway", "Compressed portraits", "Candid events"}, DOFNotes: "Extremely creamy bokeh, very shallow DOF possible", Distortion: "Pleasing compression—great face proportions"},
		{FocalLength: "70-200mm", Type: "Zoom Telephoto", BestFor: []string{"Weddings", "Events", "Versatile studio", "Fashion"}, DOFNotes: "Variable; f/2.8 constant gives good isolation", Distortion: "Slight compression, flexible framing"},
	}
}
