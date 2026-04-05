(function () {
    "use strict";

    const ROLE_COLORS = {
        key: "#FF9800", fill: "#4CAF50", rim: "#E91E63",
        hair: "#9C27B0", background: "#00BCD4", accent: "#FFC107", kicker: "#FF5722"
    };

    const MODIFIER_LABELS = {
        softbox: "Softbox", octabox: "Octabox", stripbox: "Strip Box",
        umbrella: "Umbrella", beauty_dish: "Beauty Dish", honeycomb_grid: "Honeycomb Grid",
        snoot: "Snoot", barn_doors: "Barn Doors", diffusion_panel: "Diffusion Panel",
        reflector: "Reflector", parabolic: "Parabolic", none: "None (Bare)"
    };

    const PANEL_TYPE_LABELS = {
        negative_fill: "Negative Fill (V-Flat Black)", bounce_white: "Bounce (White)",
        bounce_silver: "Bounce (Silver)", bounce_gold: "Bounce (Gold)",
        diffusion_scrim: "Diffusion Scrim", flag: "Flag / Gobo"
    };

    const PANEL_SIZE_LABELS = {
        small: "Small (12×16″)", medium: "Medium (20×30″)",
        large: "Large (4×8′ V-Flat)", xlarge: "X-Large (Scrim/Overhead)"
    };

    const PANEL_COLORS = {
        negative_fill: "#333333", bounce_white: "#E0E0E0",
        bounce_silver: "#B0B0B0", bounce_gold: "#D4A017",
        diffusion_scrim: "#F0F0F0", flag: "#1A1A1A"
    };

    const LIGHT_TYPE_LABELS = {
        speedlight: "Speedlight", strobe: "Studio Strobe",
        continuous: "Continuous", led_panel: "LED Panel",
        ring_light: "Ring Light", natural: "Natural",
        sun: "Sun (Outdoor)"
    };

    let scene = {
        id: "custom", name: "Custom Setup", mode: "portrait",
        lights: [], panels: [], backdrop: "#1a1a1a", ambient: 0.1, notes: "",
        camera: {
            focal_length: 85, aperture: 2.8, shutter_speed: "1/200",
            iso: 100, white_balance: 5500, sensor_size: "full_frame",
            angle_x: 0, angle_y: 0, distance: 2.5
        }
    };

    const CUSTOM_PRESETS_KEY = "light_sim_custom_presets";

    let lightIdCounter = 0;
    let panelIdCounter = 0;
    let dragState = null;
    let currentPresetMeta = null;

    // ---- Initialization ----

    function init() {
        loadPresets();
        bindToolbar();
        bindTabs();
        bindCameraControls();
        bindSceneControls();
        bindDiagramInteraction();
        bindSaveDialog();
        renderCustomPresetsManager();
        updatePreview();
        loadPresetFromURL();
    }

    async function loadPresetFromURL() {
        const params = new URLSearchParams(window.location.search);
        const presetId = params.get("preset");
        if (!presetId) return;
        try {
            const resp = await fetch(`/api/presets/${presetId}`);
            if (!resp.ok) return;
            const preset = await resp.json();
            applyPreset(preset);
            const select = document.getElementById("presetSelect");
            select.value = presetId;
        } catch (e) {
            console.error("Failed to load preset from URL:", e);
        }
    }

    // ---- Preset Loading ----

    async function loadPresets() {
        try {
            const resp = await fetch("/api/presets");
            const categories = await resp.json();
            const select = document.getElementById("presetSelect");
            for (const [cat, presets] of Object.entries(categories)) {
                const group = document.createElement("optgroup");
                group.label = cat.charAt(0).toUpperCase() + cat.slice(1);
                for (const p of presets) {
                    const opt = document.createElement("option");
                    opt.value = p.id;
                    opt.textContent = p.name;
                    group.appendChild(opt);
                }
                select.appendChild(group);
            }
            renderCustomPresetsInDropdown(select);

            select.addEventListener("change", async () => {
                if (!select.value) return;
                if (select.value.startsWith("custom_")) {
                    const customs = getCustomPresets();
                    const preset = customs.find(c => c.id === select.value);
                    if (preset) applyPreset(preset);
                    return;
                }
                try {
                    const r = await fetch(`/api/presets/${select.value}`);
                    const preset = await r.json();
                    applyPreset(preset);
                } catch (e) {
                    console.error("Failed to load preset:", e);
                }
            });
        } catch (e) {
            console.error("Failed to load presets:", e);
        }
    }

    function renderCustomPresetsInDropdown(select) {
        const existing = select.querySelector('optgroup[label="My Presets"]');
        if (existing) existing.remove();

        const customs = getCustomPresets();
        if (customs.length === 0) return;

        const group = document.createElement("optgroup");
        group.label = "My Presets";
        for (const c of customs) {
            const opt = document.createElement("option");
            opt.value = c.id;
            opt.textContent = c.name + " \u2605";
            group.appendChild(opt);
        }
        select.insertBefore(group, select.firstChild.nextSibling);
    }

    function applyPreset(preset) {
        currentPresetMeta = preset;
        scene.mode = preset.scene.mode;
        scene.lights = preset.scene.lights.map(l => ({ ...l }));
        scene.panels = (preset.scene.panels || []).map(p => ({ ...p, position: { ...p.position } }));
        scene.camera = { ...preset.scene.camera };
        scene.backdrop = preset.scene.backdrop;
        scene.ambient = preset.scene.ambient;

        document.getElementById("shootMode").value = scene.mode;
        syncCameraUI();
        syncSceneUI();
        renderLightsList();
        renderPanelsList();
        renderDiagram();
        updatePreview();
        showPresetDetails(preset);
    }

    function showPresetDetails(preset) {
        const panel = document.getElementById("analysisPanel");
        const content = document.getElementById("analysisContent");
        panel.hidden = false;

        let html = `<h3 class="equipment-panel__title">${preset.name}</h3>`;
        if (preset.description) {
            html += `<p class="preset-description">${preset.description}</p>`;
        }

        html += buildFlashSettingsHTML(preset.scene.lights, preset.scene.camera);

        if (preset.scene.panels && preset.scene.panels.length > 0) {
            html += buildPanelSettingsHTML(preset.scene.panels);
        }

        if (preset.equipment && preset.equipment.length > 0) {
            html += `<h4 class="flash-section__heading">Equipment & Accessories</h4>`;
            html += `<table class="equipment-panel__table">
                <thead><tr><th>Role</th><th>Device</th><th>Modifier</th><th>Power</th><th>Placement</th><th>Recommended</th></tr></thead><tbody>`;
            for (const e of preset.equipment) {
                html += `<tr><td>${e.role}</td><td>${e.device}</td><td>${e.modifier}</td><td>${e.power}</td><td>${e.placement}</td><td>${e.recommended}</td></tr>`;
            }
            html += `</tbody></table>`;
        }

        content.innerHTML = html;
    }

    function buildFlashSettingsHTML(lights, camera) {
        let html = `<h4 class="flash-section__heading">Flash & Camera Settings</h4>`;

        html += `<div class="flash-settings__camera">
            <span><strong>Camera:</strong> ${camera.focal_length}mm &middot; f/${camera.aperture} &middot; ${camera.shutter_speed}s &middot; ISO ${camera.iso} &middot; WB ${camera.white_balance}K &middot; ${camera.sensor_size.replace('_', ' ')}</span>
        </div>`;

        html += `<table class="flash-settings__table">
            <thead><tr><th>Light</th><th>Type</th><th>Modifier</th><th>Role</th><th>Power</th><th>Temp</th><th>CRI</th><th>Dist</th><th>Angle</th><th>Height</th><th>Grid</th><th>Feathered</th></tr></thead><tbody>`;

        for (const l of lights) {
            const typeLbl = LIGHT_TYPE_LABELS[l.type] || l.type;
            const modLbl = MODIFIER_LABELS[l.modifier] || l.modifier;
            const roleLbl = l.role.charAt(0).toUpperCase() + l.role.slice(1);
            const roleColor = ROLE_COLORS[l.role] || "#888";
            const dist = l.position.distance ? l.position.distance.toFixed(1) + "m" : "—";
            const angle = l.position.angle !== undefined ? Math.round(l.position.angle) + "°" : "—";
            const height = l.position.y !== undefined ? l.position.y.toFixed(1) + "m" : "—";
            const grid = l.grid_degree > 0 ? l.grid_degree + "°" : "Off";
            const feathered = l.feathered ? "Yes" : "No";

            html += `<tr>
                <td><span class="flash-role-dot" style="background:${roleColor}"></span>${l.name}</td>
                <td>${typeLbl}</td><td>${modLbl}</td>
                <td>${roleLbl}</td><td>${l.power}%</td><td>${l.color_temp}K</td>
                <td>${l.cri || "—"}</td><td>${dist}</td><td>${angle}</td>
                <td>${height}</td><td>${grid}</td><td>${feathered}</td>
            </tr>`;
        }
        html += `</tbody></table>`;
        return html;
    }

    function buildPanelSettingsHTML(panels) {
        let html = `<h4 class="flash-section__heading">Fill Panels & Modifiers</h4>`;
        html += `<table class="flash-settings__table">
            <thead><tr><th>Panel</th><th>Type</th><th>Size</th><th>Dist</th><th>Angle</th><th>Rotation</th></tr></thead><tbody>`;
        for (const p of panels) {
            const typeLbl = PANEL_TYPE_LABELS[p.type] || p.type;
            const sizeLbl = PANEL_SIZE_LABELS[p.size] || p.size;
            const panelColor = PANEL_COLORS[p.type] || "#888";
            const dist = p.position.distance ? p.position.distance.toFixed(1) + "m" : "—";
            const angle = p.position.angle !== undefined ? Math.round(p.position.angle) + "°" : "—";
            const rotation = p.rotation !== undefined ? Math.round(p.rotation) + "°" : "0°";
            html += `<tr>
                <td><span class="flash-role-dot" style="background:${panelColor}"></span>${p.name}</td>
                <td>${typeLbl}</td><td>${sizeLbl}</td>
                <td>${dist}</td><td>${angle}</td><td>${rotation}</td>
            </tr>`;
        }
        html += `</tbody></table>`;
        return html;
    }

    // ---- Custom Preset Storage (localStorage) ----

    function getCustomPresets() {
        try {
            return JSON.parse(localStorage.getItem(CUSTOM_PRESETS_KEY) || "[]");
        } catch (_) {
            return [];
        }
    }

    function saveCustomPresets(list) {
        localStorage.setItem(CUSTOM_PRESETS_KEY, JSON.stringify(list));
    }

    function addCustomPreset(name, category, notes) {
        const customs = getCustomPresets();
        const id = "custom_" + Date.now();
        const preset = {
            id, name,
            category: category || "custom",
            description: notes || "",
            equipment: [],
            scene: JSON.parse(JSON.stringify(scene)),
            diagram: ""
        };
        customs.push(preset);
        saveCustomPresets(customs);
        return preset;
    }

    function deleteCustomPreset(id) {
        const customs = getCustomPresets().filter(c => c.id !== id);
        saveCustomPresets(customs);
    }

    function renameCustomPreset(id, newName) {
        const customs = getCustomPresets();
        const p = customs.find(c => c.id === id);
        if (p) {
            p.name = newName;
            saveCustomPresets(customs);
        }
    }

    function updateCustomPreset(id) {
        const customs = getCustomPresets();
        const p = customs.find(c => c.id === id);
        if (p) {
            p.scene = JSON.parse(JSON.stringify(scene));
            saveCustomPresets(customs);
        }
    }

    // ---- Save Dialog ----

    function bindSaveDialog() {
        document.getElementById("savePresetBtn").addEventListener("click", openSaveDialog);
        document.getElementById("cancelSaveBtn").addEventListener("click", closeSaveDialog);
        document.getElementById("confirmSaveBtn").addEventListener("click", confirmSave);
        document.getElementById("saveDialogOverlay").addEventListener("click", (e) => {
            if (e.target === e.currentTarget) closeSaveDialog();
        });
    }

    function openSaveDialog() {
        const overlay = document.getElementById("saveDialogOverlay");
        const nameInput = document.getElementById("savePresetName");
        const catSelect = document.getElementById("savePresetCategory");
        nameInput.value = "";
        catSelect.value = scene.mode || "custom";
        document.getElementById("savePresetNotes").value = "";
        overlay.hidden = false;
        nameInput.focus();
    }

    function closeSaveDialog() {
        document.getElementById("saveDialogOverlay").hidden = true;
    }

    function confirmSave() {
        const name = document.getElementById("savePresetName").value.trim();
        if (!name) {
            document.getElementById("savePresetName").focus();
            return;
        }
        const category = document.getElementById("savePresetCategory").value;
        const notes = document.getElementById("savePresetNotes").value.trim();
        const preset = addCustomPreset(name, category, notes);
        closeSaveDialog();
        refreshCustomPresetsUI();
        const select = document.getElementById("presetSelect");
        select.value = preset.id;
        showPresetDetails(preset);
    }

    function refreshCustomPresetsUI() {
        const select = document.getElementById("presetSelect");
        renderCustomPresetsInDropdown(select);
        renderCustomPresetsManager();
    }

    function renderCustomPresetsManager() {
        const panel = document.getElementById("customPresetsPanel");
        const list = document.getElementById("customPresetsList");
        const customs = getCustomPresets();

        if (customs.length === 0) {
            panel.hidden = true;
            return;
        }
        panel.hidden = false;

        list.innerHTML = customs.map(c => `
            <div class="custom-preset-item" data-id="${c.id}">
                <span class="custom-preset-item__name" title="${c.description || ''}">${c.name}</span>
                <span class="custom-preset-item__cat">${c.category}</span>
                <div class="custom-preset-item__actions">
                    <button class="btn btn--xs btn--primary custom-load" title="Load">Load</button>
                    <button class="btn btn--xs btn--secondary custom-rename" title="Rename">Rename</button>
                    <button class="btn btn--xs btn--save custom-update" title="Overwrite with current settings">Update</button>
                    <button class="btn btn--xs btn--danger custom-delete" title="Delete">&times;</button>
                </div>
            </div>
        `).join("");

        list.querySelectorAll(".custom-load").forEach(btn => {
            btn.addEventListener("click", () => {
                const id = btn.closest(".custom-preset-item").dataset.id;
                const preset = getCustomPresets().find(c => c.id === id);
                if (preset) {
                    applyPreset(preset);
                    document.getElementById("presetSelect").value = id;
                }
            });
        });

        list.querySelectorAll(".custom-rename").forEach(btn => {
            btn.addEventListener("click", () => {
                const item = btn.closest(".custom-preset-item");
                const id = item.dataset.id;
                const nameSpan = item.querySelector(".custom-preset-item__name");
                const current = nameSpan.textContent;
                const newName = prompt("Rename preset:", current);
                if (newName && newName.trim()) {
                    renameCustomPreset(id, newName.trim());
                    refreshCustomPresetsUI();
                }
            });
        });

        list.querySelectorAll(".custom-update").forEach(btn => {
            btn.addEventListener("click", () => {
                const id = btn.closest(".custom-preset-item").dataset.id;
                if (confirm("Overwrite this preset with the current scene settings?")) {
                    updateCustomPreset(id);
                    refreshCustomPresetsUI();
                }
            });
        });

        list.querySelectorAll(".custom-delete").forEach(btn => {
            btn.addEventListener("click", () => {
                const id = btn.closest(".custom-preset-item").dataset.id;
                if (confirm("Delete this custom preset?")) {
                    deleteCustomPreset(id);
                    refreshCustomPresetsUI();
                    const select = document.getElementById("presetSelect");
                    if (select.value === id) select.value = "";
                }
            });
        });

        document.getElementById("closeCustomPanel").addEventListener("click", () => {
            panel.hidden = true;
        });
    }

    // ---- Toolbar ----

    function bindToolbar() {
        document.getElementById("addLightBtn").addEventListener("click", addNewLight);
        document.getElementById("addPanelBtn").addEventListener("click", addNewPanel);

        const accessorySelect = document.getElementById("accessorySelect");
        if (accessorySelect) {
            accessorySelect.addEventListener("change", (e) => {
                const val = e.target.value;
                if (!val) return;
                addAccessory(val);
                e.target.value = "";
            });
        }

        document.getElementById("uploadBtn").addEventListener("click", () => {
            document.getElementById("photoInput").click();
        });

        document.getElementById("photoInput").addEventListener("change", handlePhotoUpload);
        document.getElementById("analyzeBtn").addEventListener("click", analyzeScene);

        document.getElementById("shootMode").addEventListener("change", (e) => {
            scene.mode = e.target.value;
        });
    }

    const ACCESSORY_DEFAULTS = {
        fill_light: {
            kind: "light",
            props: { type: "strobe", modifier: "softbox", role: "fill", power: 40, color_temp: 5500, cri: 95, distance: 2.0 }
        },
        rim_light: {
            kind: "light",
            props: { type: "strobe", modifier: "stripbox", role: "rim", power: 60, color_temp: 5500, cri: 95, distance: 2.5 }
        },
        hair_light: {
            kind: "light",
            props: { type: "strobe", modifier: "snoot", role: "hair", power: 50, color_temp: 5500, cri: 95, distance: 2.0 }
        },
        bg_light: {
            kind: "light",
            props: { type: "strobe", modifier: "reflector", role: "background", power: 30, color_temp: 5500, cri: 95, distance: 3.0 }
        },
        sun: {
            kind: "light",
            props: { type: "sun", modifier: "none", role: "key", power: 100, color_temp: 5600, cri: 100, distance: 3.5 }
        },
        neg_fill: {
            kind: "panel",
            props: { type: "negative_fill", size: "large", name: "V-Flat (Black)" }
        },
        bounce_white: {
            kind: "panel",
            props: { type: "bounce_white", size: "large", name: "Bounce (White)" }
        },
        bounce_silver: {
            kind: "panel",
            props: { type: "bounce_silver", size: "medium", name: "Bounce (Silver)" }
        },
        bounce_gold: {
            kind: "panel",
            props: { type: "bounce_gold", size: "medium", name: "Bounce (Gold)" }
        },
        flag: {
            kind: "panel",
            props: { type: "flag", size: "medium", name: "Flag / Gobo" }
        },
        diffusion: {
            kind: "panel",
            props: { type: "diffusion_scrim", size: "large", name: "Diffusion Scrim" }
        }
    };

    function addAccessory(key) {
        const def = ACCESSORY_DEFAULTS[key];
        if (!def) return;

        const angle = Math.random() * 360;
        const rad = angle * Math.PI / 180;

        if (def.kind === "light") {
            lightIdCounter++;
            const p = def.props;
            const dist = p.distance || 2.0;
            scene.lights.push({
                id: `light_${lightIdCounter}`,
                name: `${p.role.charAt(0).toUpperCase() + p.role.slice(1)} ${lightIdCounter}`,
                type: p.type, modifier: p.modifier, role: p.role,
                position: {
                    x: Math.sin(rad) * dist, y: 0.5, z: Math.cos(rad) * dist,
                    distance: dist, angle
                },
                power: p.power, color_temp: p.color_temp, cri: p.cri,
                gel_color: "", grid_degree: 0, feathered: false, enabled: true
            });
            renderLightsList();
        } else {
            panelIdCounter++;
            const p = def.props;
            const dist = 1.0;
            scene.panels.push({
                id: `panel_${panelIdCounter}`,
                name: p.name || `Panel ${panelIdCounter}`,
                type: p.type, size: p.size,
                position: {
                    x: Math.sin(rad) * dist, y: 0, z: Math.cos(rad) * dist,
                    distance: dist, angle
                },
                rotation: 0, enabled: true
            });
            renderPanelsList();
        }
        renderDiagram();
        updatePreview();
    }

    function addNewLight() {
        lightIdCounter++;
        const id = `light_${lightIdCounter}`;
        const angle = Math.random() * 360;
        const rad = angle * Math.PI / 180;
        scene.lights.push({
            id, name: `Light ${lightIdCounter}`, type: "strobe",
            modifier: "softbox", role: "key",
            position: {
                x: Math.sin(rad) * 2, y: 0.5, z: Math.cos(rad) * 2,
                distance: 2.0, angle
            },
            power: 70, color_temp: 5500, cri: 95, gel_color: "",
            grid_degree: 0, feathered: false, enabled: true
        });
        renderLightsList();
        renderDiagram();
        updatePreview();
    }

    function addNewPanel() {
        panelIdCounter++;
        const id = `panel_${panelIdCounter}`;
        const angle = Math.random() * 360;
        const rad = angle * Math.PI / 180;
        scene.panels.push({
            id, name: `Panel ${panelIdCounter}`, type: "bounce_white",
            size: "medium",
            position: {
                x: Math.sin(rad) * 1.0, y: 0, z: Math.cos(rad) * 1.0,
                distance: 1.0, angle
            },
            rotation: 0, enabled: true
        });
        renderPanelsList();
        renderDiagram();
        updatePreview();
    }

    async function handlePhotoUpload(e) {
        const file = e.target.files[0];
        if (!file) return;

        const uploadBtn = document.getElementById("uploadBtn");
        const subjectImg = document.getElementById("subjectImg");
        const subjectContainer = document.getElementById("subjectContainer");

        uploadBtn.textContent = "Processing…";
        uploadBtn.disabled = true;
        subjectContainer.classList.add("preview__subject--loading");

        const formData = new FormData();
        formData.append("photo", file);

        try {
            const resp = await fetch("/api/upload", { method: "POST", body: formData });
            if (!resp.ok) {
                const err = await resp.json().catch(() => ({ error: "Upload failed" }));
                throw new Error(err.error || "Upload failed");
            }
            const data = await resp.json();
            if (data.url) {
                subjectImg.onload = function () {
                    subjectContainer.classList.remove("preview__subject--loading");
                    updatePreview();
                };
                subjectImg.onerror = function () {
                    subjectContainer.classList.remove("preview__subject--loading");
                    subjectImg.src = "/static/images/default-subject.svg";
                };
                subjectImg.src = data.url;
            }
        } catch (err) {
            console.error("Upload failed:", err);
            subjectContainer.classList.remove("preview__subject--loading");
            showUploadError(err.message);
        } finally {
            uploadBtn.textContent = "Upload Photo";
            uploadBtn.disabled = false;
            e.target.value = "";
        }
    }

    function showUploadError(msg) {
        const panel = document.getElementById("analysisPanel");
        const content = document.getElementById("analysisContent");
        panel.hidden = false;
        content.innerHTML = `<div class="analysis-warning">Upload error: ${msg}</div>`;
    }

    async function analyzeScene() {
        try {
            const resp = await fetch("/api/analyze", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(scene)
            });
            const analysis = await resp.json();
            showAnalysis(analysis);
        } catch (err) {
            console.error("Analysis failed:", err);
        }
    }

    function showAnalysis(analysis) {
        const panel = document.getElementById("analysisPanel");
        const content = document.getElementById("analysisContent");
        panel.hidden = false;

        let html = `
            <div class="analysis-item"><span>Exposure Value</span><span>${analysis.overall_ev} EV</span></div>
            <div class="analysis-item"><span>Key:Fill Ratio</span><span>${analysis.key_to_fill_ratio ? analysis.key_to_fill_ratio.toFixed(1) + ':1' : 'N/A'}</span></div>
            <div class="analysis-item"><span>Shadow Quality</span><span>${analysis.shadow_quality}</span></div>
            <div class="analysis-item"><span>Catchlight</span><span>${analysis.catchlight_type}</span></div>
        `;

        if (analysis.panel_effects && analysis.panel_effects.length > 0) {
            html += `<h4 class="flash-section__heading" style="margin-top:0.5rem">Panel Effects</h4>`;
            for (const pe of analysis.panel_effects) {
                const sign = pe.effect_intensity >= 0 ? "+" : "";
                html += `<div class="analysis-item"><span>${pe.panel_id}</span><span>${sign}${pe.effect_intensity.toFixed(1)} intensity — ${pe.description}</span></div>`;
            }
        }

        if (analysis.warnings) {
            for (const w of analysis.warnings) {
                html += `<div class="analysis-warning">${w}</div>`;
            }
        }

        content.innerHTML = html;

        if (analysis.css_filters) {
            applyCSSFilters(analysis.css_filters);
        }
    }

    // ---- Tabs ----

    function bindTabs() {
        document.querySelectorAll(".tab").forEach(tab => {
            tab.addEventListener("click", () => {
                document.querySelectorAll(".tab").forEach(t => t.classList.remove("active"));
                document.querySelectorAll(".tab-content").forEach(c => c.classList.remove("active"));
                tab.classList.add("active");
                document.getElementById(`tab-${tab.dataset.tab}`).classList.add("active");
            });
        });
    }

    // ---- Camera Controls ----

    function bindCameraControls() {
        bindRange("focalLength", "focalVal", v => {
            scene.camera.focal_length = parseInt(v);
            return v + "mm";
        });

        bindRange("aperture", "apertureVal", v => {
            const fstop = parseInt(v) / 10;
            scene.camera.aperture = fstop;
            return "f/" + fstop;
        });

        bindSelect("iso", v => { scene.camera.iso = parseInt(v); });
        bindSelect("shutterSpeed", v => { scene.camera.shutter_speed = v; });

        bindRange("whiteBalance", "wbVal", v => {
            scene.camera.white_balance = parseInt(v);
            return v + "K";
        });

        bindSelect("sensorSize", v => { scene.camera.sensor_size = v; });

        bindRange("camDistance", "camDistVal", v => {
            const d = parseInt(v) / 10;
            scene.camera.distance = d;
            updateCameraPosition();
            return d.toFixed(1) + "m";
        });

        bindRange("camAngle", "camAngleVal", v => {
            scene.camera.angle_x = parseInt(v);
            return v + "°";
        });
    }

    function syncCameraUI() {
        setVal("focalLength", scene.camera.focal_length, "focalVal", scene.camera.focal_length + "mm");
        setVal("aperture", Math.round(scene.camera.aperture * 10), "apertureVal", "f/" + scene.camera.aperture);
        document.getElementById("iso").value = scene.camera.iso;
        document.getElementById("shutterSpeed").value = scene.camera.shutter_speed;
        setVal("whiteBalance", scene.camera.white_balance, "wbVal", scene.camera.white_balance + "K");
        document.getElementById("sensorSize").value = scene.camera.sensor_size;
        setVal("camDistance", Math.round(scene.camera.distance * 10), "camDistVal", scene.camera.distance.toFixed(1) + "m");
        setVal("camAngle", scene.camera.angle_x, "camAngleVal", scene.camera.angle_x + "°");
    }

    function setVal(inputId, val, labelId, labelText) {
        document.getElementById(inputId).value = val;
        document.getElementById(labelId).textContent = labelText;
    }

    // ---- Scene Controls ----

    function bindSceneControls() {
        document.getElementById("backdropColor").addEventListener("input", (e) => {
            scene.backdrop = e.target.value;
            updatePreview();
        });

        bindRange("ambientLight", "ambientVal", v => {
            scene.ambient = parseInt(v) / 100;
            updatePreview();
            return v + "%";
        });
    }

    function syncSceneUI() {
        document.getElementById("backdropColor").value = scene.backdrop;
        setVal("ambientLight", Math.round(scene.ambient * 100), "ambientVal", Math.round(scene.ambient * 100) + "%");
    }

    // ---- Light List Rendering ----

    function renderLightsList() {
        const container = document.getElementById("lightsList");
        container.innerHTML = "";

        for (const light of scene.lights) {
            const card = document.createElement("div");
            card.className = "light-card";
            card.innerHTML = buildLightCardHTML(light);
            bindLightCard(card, light);
            container.appendChild(card);
        }
    }

    function buildLightCardHTML(light) {
        const roleColor = ROLE_COLORS[light.role] || "#888";
        return `
            <div class="light-card__header">
                <span class="light-card__name">
                    <span class="light-card__role-dot" style="background:${roleColor}"></span>
                    ${light.name}
                </span>
                <div class="light-card__actions">
                    <label style="font-size:0.75rem;display:flex;align-items:center;gap:2px">
                        <input type="checkbox" class="light-enabled" ${light.enabled ? 'checked' : ''}>On
                    </label>
                    <button class="btn btn--sm btn--danger light-delete">×</button>
                </div>
            </div>
            <div class="control-group">
                <label>Role</label>
                <select class="light-role">
                    ${["key","fill","rim","hair","background","accent","kicker"].map(r =>
                        `<option value="${r}" ${light.role === r ? 'selected' : ''}>${r.charAt(0).toUpperCase() + r.slice(1)}</option>`
                    ).join('')}
                </select>
            </div>
            <div class="control-group">
                <label>Type</label>
                <select class="light-type">
                    ${Object.entries(LIGHT_TYPE_LABELS).map(([k,v]) =>
                        `<option value="${k}" ${light.type === k ? 'selected' : ''}>${v}</option>`
                    ).join('')}
                </select>
            </div>
            <div class="control-group">
                <label>Modifier</label>
                <select class="light-modifier">
                    ${Object.entries(MODIFIER_LABELS).map(([k,v]) =>
                        `<option value="${k}" ${light.modifier === k ? 'selected' : ''}>${v}</option>`
                    ).join('')}
                </select>
            </div>
            <div class="control-group">
                <label>Power <span class="val light-power-val">${light.power}%</span></label>
                <input type="range" class="light-power" min="0" max="100" value="${light.power}">
            </div>
            <div class="control-group">
                <label>Color Temp <span class="val light-temp-val">${light.color_temp}K</span></label>
                <input type="range" class="light-temp" min="2500" max="8000" value="${light.color_temp}" step="100">
            </div>
            <div class="control-group">
                <label>Distance <span class="val light-dist-val">${light.position.distance.toFixed(1)}m</span></label>
                <input type="range" class="light-distance" min="3" max="60" value="${Math.round(light.position.distance * 10)}">
            </div>
            <div class="control-group">
                <label>Angle <span class="val light-angle-val">${Math.round(light.position.angle)}°</span></label>
                <input type="range" class="light-angle" min="0" max="360" value="${Math.round(light.position.angle)}">
            </div>
            <div class="control-group">
                <label>Height <span class="val light-height-val">${light.position.y.toFixed(1)}m</span></label>
                <input type="range" class="light-height" min="-10" max="30" value="${Math.round(light.position.y * 10)}">
            </div>
            <div class="control-group" style="display:flex;gap:0.5rem;align-items:center">
                <label style="flex:1"><input type="checkbox" class="light-feathered" ${light.feathered ? 'checked' : ''}> Feathered</label>
                <label style="font-size:0.75rem">Grid°</label>
                <select class="light-grid" style="width:60px">
                    ${[0,10,20,30,40,60].map(g => `<option value="${g}" ${light.grid_degree === g ? 'selected' : ''}>${g || 'Off'}</option>`).join('')}
                </select>
            </div>
        `;
    }

    function bindLightCard(card, light) {
        card.querySelector(".light-enabled").addEventListener("change", (e) => {
            light.enabled = e.target.checked;
            refresh();
        });

        card.querySelector(".light-delete").addEventListener("click", () => {
            scene.lights = scene.lights.filter(l => l.id !== light.id);
            renderLightsList();
            refresh();
        });

        card.querySelector(".light-role").addEventListener("change", (e) => {
            light.role = e.target.value;
            renderLightsList();
            refresh();
        });

        card.querySelector(".light-type").addEventListener("change", (e) => { light.type = e.target.value; });
        card.querySelector(".light-modifier").addEventListener("change", (e) => { light.modifier = e.target.value; refresh(); });

        bindCardRange(card, ".light-power", ".light-power-val", v => {
            light.power = parseInt(v); return v + "%";
        });
        bindCardRange(card, ".light-temp", ".light-temp-val", v => {
            light.color_temp = parseInt(v); return v + "K";
        });
        bindCardRange(card, ".light-distance", ".light-dist-val", v => {
            const d = parseInt(v) / 10;
            light.position.distance = d;
            updateLightXZ(light);
            renderDiagram();
            return d.toFixed(1) + "m";
        });
        bindCardRange(card, ".light-angle", ".light-angle-val", v => {
            light.position.angle = parseInt(v);
            updateLightXZ(light);
            renderDiagram();
            return v + "°";
        });
        bindCardRange(card, ".light-height", ".light-height-val", v => {
            light.position.y = parseInt(v) / 10;
            return (parseInt(v) / 10).toFixed(1) + "m";
        });

        card.querySelector(".light-feathered").addEventListener("change", (e) => {
            light.feathered = e.target.checked; refresh();
        });
        card.querySelector(".light-grid").addEventListener("change", (e) => {
            light.grid_degree = parseInt(e.target.value); refresh();
        });
    }

    // ---- Panel List Rendering ----

    function renderPanelsList() {
        const container = document.getElementById("panelsList");
        if (!container) return;
        container.innerHTML = "";

        for (const panel of scene.panels) {
            const card = document.createElement("div");
            card.className = "panel-card";
            card.innerHTML = buildPanelCardHTML(panel);
            bindPanelCard(card, panel);
            container.appendChild(card);
        }
    }

    function buildPanelCardHTML(panel) {
        const panelColor = PANEL_COLORS[panel.type] || "#888";
        return `
            <div class="light-card__header">
                <span class="light-card__name">
                    <span class="light-card__role-dot" style="background:${panelColor}"></span>
                    ${panel.name}
                </span>
                <div class="light-card__actions">
                    <label style="font-size:0.75rem;display:flex;align-items:center;gap:2px">
                        <input type="checkbox" class="panel-enabled" ${panel.enabled ? 'checked' : ''}>On
                    </label>
                    <button class="btn btn--sm btn--danger panel-delete">&times;</button>
                </div>
            </div>
            <div class="control-group">
                <label>Type</label>
                <select class="panel-type">
                    ${Object.entries(PANEL_TYPE_LABELS).map(([k,v]) =>
                        `<option value="${k}" ${panel.type === k ? 'selected' : ''}>${v}</option>`
                    ).join('')}
                </select>
            </div>
            <div class="control-group">
                <label>Size</label>
                <select class="panel-size">
                    ${Object.entries(PANEL_SIZE_LABELS).map(([k,v]) =>
                        `<option value="${k}" ${panel.size === k ? 'selected' : ''}>${v}</option>`
                    ).join('')}
                </select>
            </div>
            <div class="control-group">
                <label>Distance <span class="val panel-dist-val">${panel.position.distance.toFixed(1)}m</span></label>
                <input type="range" class="panel-distance" min="2" max="40" value="${Math.round(panel.position.distance * 10)}">
            </div>
            <div class="control-group">
                <label>Angle <span class="val panel-angle-val">${Math.round(panel.position.angle)}°</span></label>
                <input type="range" class="panel-angle" min="0" max="360" value="${Math.round(panel.position.angle)}">
            </div>
            <div class="control-group">
                <label>Rotation <span class="val panel-rot-val">${Math.round(panel.rotation)}°</span></label>
                <input type="range" class="panel-rotation" min="0" max="360" value="${Math.round(panel.rotation)}">
            </div>
        `;
    }

    function bindPanelCard(card, panel) {
        card.querySelector(".panel-enabled").addEventListener("change", (e) => {
            panel.enabled = e.target.checked;
            refresh();
        });

        card.querySelector(".panel-delete").addEventListener("click", () => {
            scene.panels = scene.panels.filter(p => p.id !== panel.id);
            renderPanelsList();
            refresh();
        });

        card.querySelector(".panel-type").addEventListener("change", (e) => {
            panel.type = e.target.value;
            renderPanelsList();
            refresh();
        });

        card.querySelector(".panel-size").addEventListener("change", (e) => {
            panel.size = e.target.value;
            refresh();
        });

        bindCardRange(card, ".panel-distance", ".panel-dist-val", v => {
            const d = parseInt(v) / 10;
            panel.position.distance = d;
            updatePanelXZ(panel);
            renderDiagram();
            return d.toFixed(1) + "m";
        });

        bindCardRange(card, ".panel-angle", ".panel-angle-val", v => {
            panel.position.angle = parseInt(v);
            updatePanelXZ(panel);
            renderDiagram();
            return v + "°";
        });

        bindCardRange(card, ".panel-rotation", ".panel-rot-val", v => {
            panel.rotation = parseInt(v);
            renderDiagram();
            return v + "°";
        });
    }

    function updatePanelXZ(panel) {
        const rad = panel.position.angle * Math.PI / 180;
        panel.position.x = Math.sin(rad) * panel.position.distance;
        panel.position.z = Math.cos(rad) * panel.position.distance;
    }

    function updateLightXZ(light) {
        const rad = light.position.angle * Math.PI / 180;
        light.position.x = Math.sin(rad) * light.position.distance;
        light.position.z = Math.cos(rad) * light.position.distance;
    }

    // ---- Diagram Rendering ----

    function renderDiagram() {
        const group = document.getElementById("lightsGroup");
        group.innerHTML = "";

        const svg = document.getElementById("diagramSvg");
        let defsEl = svg.querySelector("defs");
        defsEl.querySelectorAll("[data-beam-grad]").forEach(el => el.remove());

        const scale = 75; // 1m = 75px in SVG

        for (const light of scene.lights) {
            if (!light.enabled) continue;
            const isSun = light.type === "sun";
            const color = isSun ? "#FFA726" : (ROLE_COLORS[light.role] || "#888");
            const svgX = light.position.x * scale;
            const svgY = -light.position.z * scale;

            const g = document.createElementNS("http://www.w3.org/2000/svg", "g");
            g.setAttribute("transform", `translate(${svgX}, ${svgY})`);
            g.setAttribute("data-light-id", light.id);
            g.style.cursor = "grab";

            const distToSubject = Math.sqrt(svgX * svgX + svgY * svgY);
            const beamSpread = isSun ? 180 : getBeamSpread(light);
            const halfAngle = (beamSpread / 2) * Math.PI / 180;
            const aimAngle = Math.atan2(-svgY, -svgX);
            const dist = Math.max(light.position.distance, 0.1);

            // Effective intensity at the subject.  Sun doesn't use inverse-square;
            // studio lights do.  This is the *real* illuminance arriving at the
            // subject after all physics (distance, type, modifier) are applied.
            let intensity;
            if (isSun) {
                const elevation = Math.max(light.position.y || 0.5, 0.5);
                const elevFactor = Math.min(elevation / 3.0, 1.0);
                intensity = light.power * 1.2 * (0.3 + 0.7 * elevFactor);
            } else {
                intensity = light.power / (dist * dist);
            }
            const normalizedIntensity = Math.min(intensity / 80, 1);

            // Effective beam reach: the cone extends toward the subject only
            // as far as the light has meaningful intensity.  At low power the
            // beam fades out before reaching the subject; at high power it
            // extends all the way.  We define "effective reach" as the
            // distance at which intensity drops below a perceptual threshold
            // (2 lux equiv.).  For inverse-square: r_eff = sqrt(P / threshold).
            // For the sun (parallel rays) the beam always reaches the subject.
            const maxReach = Math.max(distToSubject - 20, 15);
            let effectiveReachPx;
            if (isSun) {
                effectiveReachPx = maxReach;
            } else {
                const reachMeters = Math.sqrt(light.power / 2);
                effectiveReachPx = Math.min(reachMeters * scale, maxReach);
                effectiveReachPx = Math.max(effectiveReachPx, 15);
            }

            const coneLen = effectiveReachPx;
            const x1 = Math.cos(aimAngle - halfAngle) * coneLen;
            const y1 = Math.sin(aimAngle - halfAngle) * coneLen;
            const x2 = Math.cos(aimAngle + halfAngle) * coneLen;
            const y2 = Math.sin(aimAngle + halfAngle) * coneLen;

            if (isSun) {
                // Sun: render parallel rays instead of a cone
                renderSunBeam(g, aimAngle, coneLen, light.power, color);
            } else {
                // Studio light: gradient cone with inverse-square falloff
                const gradId = "beamGrad_" + light.id;
                const grad = document.createElementNS("http://www.w3.org/2000/svg", "linearGradient");
                grad.setAttribute("id", gradId);
                grad.setAttribute("data-beam-grad", "1");
                grad.setAttribute("gradientUnits", "userSpaceOnUse");
                grad.setAttribute("x1", "0");
                grad.setAttribute("y1", "0");
                grad.setAttribute("x2", Math.cos(aimAngle) * coneLen);
                grad.setAttribute("y2", Math.sin(aimAngle) * coneLen);

                const peakOpacity = 0.08 + normalizedIntensity * 0.32;
                const falloffStops = [
                    { offset: "0%",   frac: 0    },
                    { offset: "15%",  frac: 0.15 },
                    { offset: "35%",  frac: 0.35 },
                    { offset: "55%",  frac: 0.55 },
                    { offset: "75%",  frac: 0.75 },
                    { offset: "100%", frac: 1.0  }
                ];

                for (const s of falloffStops) {
                    const stop = document.createElementNS("http://www.w3.org/2000/svg", "stop");
                    stop.setAttribute("offset", s.offset);
                    const relIntensity = 1 / ((1 + 2 * s.frac) * (1 + 2 * s.frac));
                    const opacity = peakOpacity * relIntensity;
                    stop.setAttribute("stop-color", color);
                    stop.setAttribute("stop-opacity", opacity.toFixed(4));
                    grad.appendChild(stop);
                }
                defsEl.appendChild(grad);

                const cone = document.createElementNS("http://www.w3.org/2000/svg", "path");
                cone.setAttribute("d", `M0,0 L${x1},${y1} L${x2},${y2} Z`);
                cone.setAttribute("fill", `url(#${gradId})`);
                cone.setAttribute("stroke", color);
                cone.setAttribute("stroke-opacity", (0.1 + normalizedIntensity * 0.25).toFixed(3));
                cone.setAttribute("stroke-width", "0.5");
                g.appendChild(cone);

                // Iso-lines at 25%, 50%, 75% of effective reach showing intensity
                const isoFractions = [0.25, 0.5, 0.75];
                for (const frac of isoFractions) {
                    const r = coneLen * frac;
                    const aStart = aimAngle - halfAngle * (1 - frac * 0.3);
                    const aEnd = aimAngle + halfAngle * (1 - frac * 0.3);
                    const ax1 = Math.cos(aStart) * r;
                    const ay1 = Math.sin(aStart) * r;
                    const ax2 = Math.cos(aEnd) * r;
                    const ay2 = Math.sin(aEnd) * r;
                    const largeArc = (aEnd - aStart) > Math.PI ? 1 : 0;

                    const arc = document.createElementNS("http://www.w3.org/2000/svg", "path");
                    arc.setAttribute("d", `M${ax1},${ay1} A${r},${r} 0 ${largeArc} 1 ${ax2},${ay2}`);
                    arc.setAttribute("fill", "none");
                    arc.setAttribute("stroke", color);
                    const arcRelI = 1 / ((1 + 2 * frac) * (1 + 2 * frac));
                    arc.setAttribute("stroke-opacity", (0.08 + arcRelI * normalizedIntensity * 0.3).toFixed(3));
                    arc.setAttribute("stroke-width", "0.4");
                    arc.setAttribute("stroke-dasharray", "2,3");
                    g.appendChild(arc);

                    const labelAngle = aimAngle + halfAngle * 0.6;
                    const lx = Math.cos(labelAngle) * r;
                    const ly = Math.sin(labelAngle) * r;
                    const isoIntensity = light.power / (Math.max(dist * frac, 0.1) ** 2);
                    const isoLabel = document.createElementNS("http://www.w3.org/2000/svg", "text");
                    isoLabel.setAttribute("x", lx);
                    isoLabel.setAttribute("y", ly);
                    isoLabel.setAttribute("text-anchor", "middle");
                    isoLabel.setAttribute("fill", color);
                    isoLabel.setAttribute("fill-opacity", (0.3 + arcRelI * 0.4).toFixed(2));
                    isoLabel.setAttribute("font-size", "6");
                    isoLabel.textContent = Math.round(isoIntensity);
                    g.appendChild(isoLabel);
                }
            }

            // Light body marker
            if (isSun) {
                renderSunMarker(g, light.power, color);
            } else {
                const circle = document.createElementNS("http://www.w3.org/2000/svg", "circle");
                circle.setAttribute("r", "12");
                circle.setAttribute("fill", color);
                circle.setAttribute("fill-opacity", (0.5 + normalizedIntensity * 0.4).toFixed(2));
                circle.setAttribute("stroke", color);
                circle.setAttribute("stroke-width", "1.5");
                g.appendChild(circle);
            }

            // Power indicator ring
            const powerRing = document.createElementNS("http://www.w3.org/2000/svg", "circle");
            powerRing.setAttribute("r", (12 + light.power * 0.08).toFixed(1));
            powerRing.setAttribute("fill", "none");
            powerRing.setAttribute("stroke", color);
            powerRing.setAttribute("stroke-opacity", (0.15 + normalizedIntensity * 0.2).toFixed(2));
            powerRing.setAttribute("stroke-width", "0.5");
            powerRing.setAttribute("stroke-dasharray", "2,2");
            g.appendChild(powerRing);

            // Label
            const text = document.createElementNS("http://www.w3.org/2000/svg", "text");
            text.setAttribute("y", "-16");
            text.setAttribute("text-anchor", "middle");
            text.setAttribute("fill", color);
            text.setAttribute("font-size", "8");
            text.textContent = light.name + " (" + light.power + "%)";
            g.appendChild(text);

            // Delete button
            const delG = createSVGDeleteButton(16, -8);
            delG.setAttribute("data-delete-light", light.id);
            g.appendChild(delG);

            group.appendChild(g);
        }

        for (const panel of scene.panels) {
            if (!panel.enabled) continue;
            const color = PANEL_COLORS[panel.type] || "#888";
            const svgX = panel.position.x * scale;
            const svgY = -panel.position.z * scale;

            const g = document.createElementNS("http://www.w3.org/2000/svg", "g");
            g.setAttribute("transform", `translate(${svgX}, ${svgY})`);
            g.setAttribute("data-panel-id", panel.id);
            g.style.cursor = "grab";

            const sizeW = getPanelSVGSize(panel.size);
            const panelGroup = document.createElementNS("http://www.w3.org/2000/svg", "g");
            if (panel.rotation) {
                panelGroup.setAttribute("transform", `rotate(${panel.rotation})`);
            }

            const rect = document.createElementNS("http://www.w3.org/2000/svg", "rect");
            rect.setAttribute("x", -sizeW / 2);
            rect.setAttribute("y", -4);
            rect.setAttribute("width", sizeW);
            rect.setAttribute("height", 8);
            rect.setAttribute("rx", "2");

            if (panel.type === "diffusion_scrim") {
                rect.setAttribute("fill", "rgba(255,255,255,0.3)");
                rect.setAttribute("stroke", "#ccc");
                rect.setAttribute("stroke-width", "1");
                rect.setAttribute("stroke-dasharray", "3,2");
            } else if (panel.type === "negative_fill" || panel.type === "flag") {
                rect.setAttribute("fill", color);
                rect.setAttribute("fill-opacity", "0.85");
                rect.setAttribute("stroke", "#666");
                rect.setAttribute("stroke-width", "1");
            } else {
                rect.setAttribute("fill", color);
                rect.setAttribute("fill-opacity", "0.6");
                rect.setAttribute("stroke", color);
                rect.setAttribute("stroke-width", "1");
            }
            panelGroup.appendChild(rect);

            // Rotation handle: small arc indicator showing panel facing direction
            const handleLen = sizeW / 2 + 6;
            const normalLine = document.createElementNS("http://www.w3.org/2000/svg", "line");
            normalLine.setAttribute("x1", "0");
            normalLine.setAttribute("y1", "-4");
            normalLine.setAttribute("x2", "0");
            normalLine.setAttribute("y2", -handleLen);
            normalLine.setAttribute("stroke", color);
            normalLine.setAttribute("stroke-opacity", "0.4");
            normalLine.setAttribute("stroke-width", "1");
            normalLine.setAttribute("stroke-dasharray", "2,2");
            panelGroup.appendChild(normalLine);

            const rotHandle = document.createElementNS("http://www.w3.org/2000/svg", "circle");
            rotHandle.setAttribute("cx", "0");
            rotHandle.setAttribute("cy", -handleLen);
            rotHandle.setAttribute("r", "4");
            rotHandle.setAttribute("fill", color);
            rotHandle.setAttribute("fill-opacity", "0.3");
            rotHandle.setAttribute("stroke", color);
            rotHandle.setAttribute("stroke-opacity", "0.6");
            rotHandle.setAttribute("stroke-width", "0.5");
            rotHandle.style.cursor = "crosshair";
            rotHandle.classList.add("panel-rotate-handle");
            panelGroup.appendChild(rotHandle);

            g.appendChild(panelGroup);

            // Label (outside the rotation group so it stays upright)
            const text = document.createElementNS("http://www.w3.org/2000/svg", "text");
            text.setAttribute("y", "-" + (sizeW / 2 + 14));
            text.setAttribute("text-anchor", "middle");
            text.setAttribute("fill", panel.type === "negative_fill" || panel.type === "flag" ? "#aaa" : "#666");
            text.setAttribute("font-size", "7");
            text.textContent = panel.name + " " + Math.round(panel.rotation || 0) + "°";
            g.appendChild(text);

            const delGP = createSVGDeleteButton(sizeW / 2 + 4, -6);
            delGP.setAttribute("data-delete-panel", panel.id);
            g.appendChild(delGP);

            group.appendChild(g);
        }

        renderPanelLightInteractions(group);
        updateCameraPosition();
    }

    function renderSunBeam(g, aimAngle, coneLen, power, color) {
        const nIntensity = Math.min(power / 80, 1);
        const rayCount = 7;
        const spread = 25;
        for (let i = 0; i < rayCount; i++) {
            const offset = ((i / (rayCount - 1)) - 0.5) * spread;
            const perpX = -Math.sin(aimAngle) * offset;
            const perpY = Math.cos(aimAngle) * offset;

            const ray = document.createElementNS("http://www.w3.org/2000/svg", "line");
            ray.setAttribute("x1", perpX);
            ray.setAttribute("y1", perpY);
            ray.setAttribute("x2", perpX + Math.cos(aimAngle) * coneLen);
            ray.setAttribute("y2", perpY + Math.sin(aimAngle) * coneLen);
            ray.setAttribute("stroke", color);
            ray.setAttribute("stroke-opacity", (0.08 + nIntensity * 0.18).toFixed(3));
            ray.setAttribute("stroke-width", "1.2");
            g.appendChild(ray);
        }

        const arrow = document.createElementNS("http://www.w3.org/2000/svg", "path");
        const tipX = Math.cos(aimAngle) * coneLen;
        const tipY = Math.sin(aimAngle) * coneLen;
        const backLen = 8;
        const backAngle = 0.4;
        const la = aimAngle + Math.PI - backAngle;
        const ra = aimAngle + Math.PI + backAngle;
        arrow.setAttribute("d",
            `M${tipX},${tipY} L${tipX + Math.cos(la) * backLen},${tipY + Math.sin(la) * backLen} ` +
            `M${tipX},${tipY} L${tipX + Math.cos(ra) * backLen},${tipY + Math.sin(ra) * backLen}`);
        arrow.setAttribute("stroke", color);
        arrow.setAttribute("stroke-opacity", (0.15 + nIntensity * 0.25).toFixed(3));
        arrow.setAttribute("stroke-width", "1.5");
        arrow.setAttribute("fill", "none");
        g.appendChild(arrow);
    }

    function renderSunMarker(g, power, color) {
        const nI = Math.min(power / 80, 1);
        const body = document.createElementNS("http://www.w3.org/2000/svg", "circle");
        body.setAttribute("r", "10");
        body.setAttribute("fill", color);
        body.setAttribute("fill-opacity", (0.7 + nI * 0.25).toFixed(2));
        body.setAttribute("stroke", "#FFF176");
        body.setAttribute("stroke-width", "1.5");
        g.appendChild(body);

        const rayLen = 5;
        for (let i = 0; i < 8; i++) {
            const a = (i / 8) * Math.PI * 2;
            const line = document.createElementNS("http://www.w3.org/2000/svg", "line");
            line.setAttribute("x1", Math.cos(a) * 12);
            line.setAttribute("y1", Math.sin(a) * 12);
            line.setAttribute("x2", Math.cos(a) * (12 + rayLen));
            line.setAttribute("y2", Math.sin(a) * (12 + rayLen));
            line.setAttribute("stroke", "#FFF176");
            line.setAttribute("stroke-width", "1");
            line.setAttribute("stroke-opacity", "0.6");
            g.appendChild(line);
        }
    }

    function renderPanelLightInteractions(group) {
        const scale = 75;
        for (const panel of scene.panels) {
            if (!panel.enabled) continue;

            const panelAngle = (panel.position.angle || 0) * Math.PI / 180;
            const panelDist = Math.max(panel.position.distance || 1, 0.1);
            const panelSvgX = Math.sin(panelAngle) * panelDist * scale;
            const panelSvgY = -(Math.cos(panelAngle) * panelDist) * scale;

            const isBounce = panel.type === "bounce_white" || panel.type === "bounce_silver" || panel.type === "bounce_gold";
            const isAbsorb = panel.type === "negative_fill" || panel.type === "flag";

            for (const light of scene.lights) {
                if (!light.enabled) continue;

                const lAngle = (light.position.angle || 0) * Math.PI / 180;
                const lDist = Math.max(light.position.distance || 1, 0.1);
                const lightSvgX = Math.sin(lAngle) * lDist * scale;
                const lightSvgY = -(Math.cos(lAngle) * lDist) * scale;

                const dx = panelSvgX - lightSvgX;
                const dy = panelSvgY - lightSvgY;
                const dist = Math.sqrt(dx * dx + dy * dy);
                if (dist < 5) continue;

                const aimDist = Math.sqrt(lightSvgX * lightSvgX + lightSvgY * lightSvgY) || 1;
                const aimX = -lightSvgX / aimDist;
                const aimY = -lightSvgY / aimDist;
                const dirX = dx / dist;
                const dirY = dy / dist;
                const cosAim = dirX * aimX + dirY * aimY;
                const aimAngleDeg = Math.acos(Math.max(-1, Math.min(1, cosAim))) * 180 / Math.PI;
                const spillHalf = getBeamSpread(light) / 2;
                if (aimAngleDeg > spillHalf) continue;

                const color = ROLE_COLORS[light.role] || "#888";

                if (isBounce) {
                    const ray = document.createElementNS("http://www.w3.org/2000/svg", "line");
                    ray.setAttribute("x1", lightSvgX);
                    ray.setAttribute("y1", lightSvgY);
                    ray.setAttribute("x2", panelSvgX);
                    ray.setAttribute("y2", panelSvgY);
                    ray.setAttribute("stroke", color);
                    ray.setAttribute("stroke-opacity", "0.12");
                    ray.setAttribute("stroke-width", "1");
                    ray.setAttribute("stroke-dasharray", "4,3");
                    group.appendChild(ray);

                    const bounceRay = document.createElementNS("http://www.w3.org/2000/svg", "line");
                    bounceRay.setAttribute("x1", panelSvgX);
                    bounceRay.setAttribute("y1", panelSvgY);
                    bounceRay.setAttribute("x2", "0");
                    bounceRay.setAttribute("y2", "0");
                    const bounceColor = panel.type === "bounce_gold" ? "#D4A017" : (panel.type === "bounce_silver" ? "#B0B0B0" : "#E0E0E0");
                    bounceRay.setAttribute("stroke", bounceColor);
                    bounceRay.setAttribute("stroke-opacity", "0.25");
                    bounceRay.setAttribute("stroke-width", "1.5");
                    bounceRay.setAttribute("stroke-dasharray", "6,3");
                    group.appendChild(bounceRay);
                } else if (isAbsorb) {
                    const ray = document.createElementNS("http://www.w3.org/2000/svg", "line");
                    ray.setAttribute("x1", lightSvgX);
                    ray.setAttribute("y1", lightSvgY);
                    ray.setAttribute("x2", panelSvgX);
                    ray.setAttribute("y2", panelSvgY);
                    ray.setAttribute("stroke", "#c62828");
                    ray.setAttribute("stroke-opacity", "0.08");
                    ray.setAttribute("stroke-width", "1");
                    ray.setAttribute("stroke-dasharray", "2,4");
                    group.appendChild(ray);
                }
            }
        }
    }

    function createSVGDeleteButton(x, y) {
        const g = document.createElementNS("http://www.w3.org/2000/svg", "g");
        g.setAttribute("transform", `translate(${x}, ${y})`);
        g.style.cursor = "pointer";
        g.classList.add("svg-delete-btn");

        const bg = document.createElementNS("http://www.w3.org/2000/svg", "circle");
        bg.setAttribute("r", "6");
        bg.setAttribute("fill", "#c62828");
        bg.setAttribute("fill-opacity", "0.85");
        bg.setAttribute("stroke", "#fff");
        bg.setAttribute("stroke-width", "0.5");
        g.appendChild(bg);

        const cross = document.createElementNS("http://www.w3.org/2000/svg", "text");
        cross.setAttribute("text-anchor", "middle");
        cross.setAttribute("y", "3");
        cross.setAttribute("fill", "#fff");
        cross.setAttribute("font-size", "8");
        cross.setAttribute("font-weight", "bold");
        cross.setAttribute("pointer-events", "none");
        cross.textContent = "\u00d7";
        g.appendChild(cross);

        return g;
    }

    function getPanelSVGSize(size) {
        switch (size) {
            case "small": return 16;
            case "medium": return 24;
            case "large": return 36;
            case "xlarge": return 48;
            default: return 24;
        }
    }

    function getBeamSpread(light) {
        if (light.modifier === "honeycomb_grid" && light.grid_degree > 0) return light.grid_degree;
        const spreads = {
            none: 120, snoot: 15, barn_doors: 40, honeycomb_grid: 30,
            reflector: 70, beauty_dish: 60, umbrella: 100, softbox: 75,
            stripbox: 50, octabox: 85, diffusion_panel: 110, parabolic: 45
        };
        return spreads[light.modifier] || 75;
    }

    function updateCameraPosition() {
        const cam = document.getElementById("cameraMarker");
        const dist = scene.camera.distance * 75;
        cam.setAttribute("transform", `translate(0, ${dist})`);
    }

    // ---- Diagram Drag Interaction ----

    function bindDiagramInteraction() {
        const svg = document.getElementById("diagramSvg");

        svg.addEventListener("click", (e) => {
            const delLight = e.target.closest("[data-delete-light]");
            if (delLight) {
                const id = delLight.getAttribute("data-delete-light");
                scene.lights = scene.lights.filter(l => l.id !== id);
                renderLightsList();
                refresh();
                e.stopPropagation();
                return;
            }
            const delPanel = e.target.closest("[data-delete-panel]");
            if (delPanel) {
                const id = delPanel.getAttribute("data-delete-panel");
                scene.panels = scene.panels.filter(p => p.id !== id);
                renderPanelsList();
                refresh();
                e.stopPropagation();
                return;
            }
        });

        svg.addEventListener("mousedown", (e) => {
            if (e.target.closest(".svg-delete-btn")) return;

            // Check for panel rotation handle first
            const rotHandle = e.target.closest(".panel-rotate-handle");
            if (rotHandle) {
                const panelG = rotHandle.closest("[data-panel-id]");
                if (panelG) {
                    const panelId = panelG.getAttribute("data-panel-id");
                    const panel = scene.panels.find(p => p.id === panelId);
                    if (panel) {
                        dragState = { item: panel, kind: "panel-rotate" };
                        e.preventDefault();
                        return;
                    }
                }
            }

            const lightG = e.target.closest("[data-light-id]");
            const panelG = e.target.closest("[data-panel-id]");

            if (lightG) {
                const lightId = lightG.getAttribute("data-light-id");
                const light = scene.lights.find(l => l.id === lightId);
                if (!light) return;
                dragState = { item: light, kind: "light" };
                lightG.style.cursor = "grabbing";
                e.preventDefault();
            } else if (panelG) {
                const panelId = panelG.getAttribute("data-panel-id");
                const panel = scene.panels.find(p => p.id === panelId);
                if (!panel) return;
                // Shift+drag rotates the panel; normal drag moves it
                if (e.shiftKey) {
                    dragState = { item: panel, kind: "panel-rotate" };
                } else {
                    dragState = { item: panel, kind: "panel" };
                }
                panelG.style.cursor = e.shiftKey ? "crosshair" : "grabbing";
                e.preventDefault();
            }
        });

        svg.addEventListener("mousemove", (e) => {
            if (!dragState) return;
            const svgRect = svg.getBoundingClientRect();
            const svgW = svgRect.width;
            const svgH = svgRect.height;

            const mx = ((e.clientX - svgRect.left) / svgW) * 600 - 300;
            const my = ((e.clientY - svgRect.top) / svgH) * 600 - 300;

            if (dragState.kind === "panel-rotate") {
                // Compute rotation angle from the panel's center to cursor
                const scale = 75;
                const panel = dragState.item;
                const panelSvgX = panel.position.x * scale;
                const panelSvgY = -panel.position.z * scale;
                const dx = mx - panelSvgX;
                const dy = my - panelSvgY;
                let angle = Math.atan2(dx, -dy) * 180 / Math.PI;
                if (angle < 0) angle += 360;
                panel.rotation = Math.round(angle);
                renderDiagram();
                renderPanelsList();
                updatePreview();
                return;
            }

            const scale = 75;
            const pos = dragState.item.position;
            pos.x = mx / scale;
            pos.z = -my / scale;
            pos.distance = Math.sqrt(pos.x ** 2 + pos.z ** 2);
            pos.angle = Math.atan2(pos.x, pos.z) * 180 / Math.PI;
            if (pos.angle < 0) pos.angle += 360;

            renderDiagram();
            if (dragState.kind === "light") renderLightsList();
            else renderPanelsList();
            updatePreview();
        });

        svg.addEventListener("mouseup", () => { dragState = null; });
        svg.addEventListener("mouseleave", () => { dragState = null; });
    }

    // ---- Preview Rendering ----

    function updatePreview() {
        const backdrop = document.getElementById("previewBackdrop");
        backdrop.style.background = scene.backdrop;

        // Compute client-side lighting approximation
        const filters = computeLocalFilters();
        applyCSSFilters(filters);
    }

    function computeLocalFilters() {
        let totalIntensity = 0;
        let weightedTemp = 0;
        let keyDir = 0;
        let keyIntensity = 0;
        let fillIntensity = 0;
        let avgSoftness = 0;
        let enabledCount = 0;

        const lightContribs = [];

        for (const light of scene.lights) {
            if (!light.enabled) continue;
            enabledCount++;

            const dist = Math.max(light.position.distance, 0.1);
            let intensity;
            if (light.type === "sun") {
                const elevation = Math.max(light.position.y || 0.5, 0.5);
                const elevFactor = Math.min(elevation / 3.0, 1.0);
                intensity = light.power * 1.2 * (0.3 + 0.7 * elevFactor);
            } else {
                intensity = light.power / (dist * dist);
            }
            totalIntensity += intensity;
            weightedTemp += light.color_temp * intensity;

            const softness = light.type === "sun" ? 0.15 : getModifierSoftness(light.modifier);
            avgSoftness += softness;

            lightContribs.push({
                intensity, softness, light,
                spillHalf: light.type === "sun" ? 90 : getBeamSpread(light) / 2
            });

            if (light.role === "key") {
                keyIntensity = intensity;
                keyDir = light.position.angle;
            }
            if (light.role === "fill") fillIntensity = intensity;
        }

        if (enabledCount > 0) avgSoftness /= enabledCount;

        // Compute panel effects using the same physics as the Go engine:
        // per-panel incident light calculation with geometric ray tracing,
        // then reflectivity/absorption applied.
        const panelEffects = computeLocalPanelEffects(lightContribs);
        let panelIntensityDelta = 0;
        let panelTempShift = 0;
        let panelSoftnessAdj = 0;

        for (const pe of panelEffects) {
            panelIntensityDelta += pe.effectIntensity;
            if (pe.colorTempShift && pe.effectIntensity > 0) {
                panelTempShift += pe.colorTempShift * pe.effectIntensity;
            }
            if (pe.effectIntensity > 0) {
                fillIntensity += pe.effectIntensity;
            } else if (pe.effectIntensity < 0) {
                fillIntensity = Math.max(0, fillIntensity + pe.effectIntensity);
            }
            if (pe.softnessModifier > 0.8) panelSoftnessAdj += 0.1;
            if (pe.effectIntensity < 0 && pe.softnessModifier === 0) panelSoftnessAdj -= 0.05;
        }

        totalIntensity = Math.max(0, totalIntensity + panelIntensityDelta);

        const filters = { brightness: 1.0, contrast: 1.0, hue_rotate: 0, warmth_shift: 0 };

        if (totalIntensity > 0) {
            if (panelTempShift !== 0) weightedTemp += panelTempShift;
            const avgTemp = weightedTemp / Math.max(totalIntensity, 0.001);
            const normalizedIntensity = Math.min(totalIntensity / 50, 1);
            filters.brightness = 0.3 + normalizedIntensity * 1.5;
            filters.warmth_shift = (avgTemp - 5500) / 3000;
            filters.hue_rotate = filters.warmth_shift > 0
                ? -filters.warmth_shift * 15
                : -filters.warmth_shift * 20;
        }

        if (fillIntensity > 0) {
            const ratio = keyIntensity / fillIntensity;
            filters.contrast = 1.0 + Math.min(ratio / 8, 0.5);
        } else if (keyIntensity > 0) {
            filters.contrast = 1.0 + Math.min((keyIntensity * 16) / 8, 0.5);
        }

        const shadowAngle = (keyDir + 180) % 360;
        let shadowQ = avgSoftness > 0.6 ? "soft" : avgSoftness > 0.3 ? "medium" : "hard";
        if (panelSoftnessAdj > 0.05 && shadowQ === "hard") shadowQ = "medium";
        else if (panelSoftnessAdj < -0.03 && shadowQ === "soft") shadowQ = "medium";

        let shadowOpacity = 0.15;
        if (shadowQ === "hard") shadowOpacity = 0.6;
        else if (shadowQ === "medium") shadowOpacity = 0.35;

        filters.shadow_gradient = `linear-gradient(${shadowAngle}deg, rgba(0,0,0,${shadowOpacity}) 0%, rgba(0,0,0,0) 60%)`;

        const hlX = 50 + Math.sin(keyDir * Math.PI / 180) * 30;
        const hlY = 50 - Math.cos(keyDir * Math.PI / 180) * 30;
        filters.highlight_pos = `radial-gradient(circle at ${hlX}% ${hlY}%, rgba(255,255,255,0.15) 0%, rgba(255,255,255,0) 50%)`;

        const evDisplay = document.getElementById("evDisplay");
        const ratioDisplay = document.getElementById("ratioDisplay");
        const shadowDisplay = document.getElementById("shadowDisplay");
        const ratio = fillIntensity > 0 ? (keyIntensity / fillIntensity).toFixed(1) : "N/A";

        evDisplay.textContent = `EV: ${(Math.log2(totalIntensity + 1) + 8).toFixed(1)}`;
        ratioDisplay.textContent = `Ratio: ${ratio}${ratio !== "N/A" ? ":1" : ""}`;
        shadowDisplay.textContent = `Shadows: ${shadowQ}`;

        return filters;
    }

    function computeLocalPanelEffects(lightContribs) {
        const effects = [];
        for (const panel of scene.panels) {
            if (!panel.enabled) continue;

            const panelAngle = (panel.position.angle || 0) * Math.PI / 180;
            const panelDist = Math.max(panel.position.distance || 1, 0.1);
            const panelX = Math.sin(panelAngle) * panelDist;
            const panelZ = Math.cos(panelAngle) * panelDist;

            const normalAngle = ((panel.position.angle || 0) + 180 + (panel.rotation || 0)) * Math.PI / 180;
            const panelNX = Math.sin(normalAngle);
            const panelNZ = Math.cos(normalAngle);

            let incidentLight = 0;
            for (const lc of lightContribs) {
                const l = lc.light;

                if (l.type === "sun") {
                    const sunAngle = (l.position.angle || 0) * Math.PI / 180;
                    const sunDirX = -Math.sin(sunAngle);
                    const sunDirZ = -Math.cos(sunAngle);
                    const cosIncidence = Math.abs(sunDirX * panelNX + sunDirZ * panelNZ);
                    incidentLight += lc.intensity * cosIncidence;
                    continue;
                }

                const lAngle = (l.position.angle || 0) * Math.PI / 180;
                const lDist = Math.max(l.position.distance || 1, 0.1);
                const lightX = Math.sin(lAngle) * lDist;
                const lightZ = Math.cos(lAngle) * lDist;

                const dx = panelX - lightX;
                const dz = panelZ - lightZ;
                let distLP = Math.sqrt(dx * dx + dz * dz);
                if (distLP < 0.05) distLP = 0.05;

                const dirX = dx / distLP;
                const dirZ = dz / distLP;

                const aimDist = Math.sqrt(lightX * lightX + lightZ * lightZ) || 0.01;
                const aimX = -lightX / aimDist;
                const aimZ = -lightZ / aimDist;
                const cosAim = dirX * aimX + dirZ * aimZ;
                const aimAngleDeg = Math.acos(Math.max(-1, Math.min(1, cosAim))) * 180 / Math.PI;

                if (aimAngleDeg > lc.spillHalf) continue;

                const cosIncidence = Math.abs(dirX * panelNX + dirZ * panelNZ);
                const spillFraction = aimAngleDeg / lc.spillHalf;
                const edgeFalloff = 1.0 - spillFraction * spillFraction;

                incidentLight += l.power * cosIncidence / (distLP * distLP) * edgeFalloff;
            }

            if (incidentLight < 0.5 && lightContribs.length > 0) {
                let totalScene = 0;
                for (const lc of lightContribs) totalScene += lc.intensity;
                incidentLight = Math.max(incidentLight, totalScene * 0.05);
            }

            const sizeFactor = getPanelSizeFactor(panel.size);
            const effect = computeLocalSinglePanelEffect(panel, sizeFactor, panelDist, incidentLight);
            effects.push(effect);
        }
        return effects;
    }

    function getPanelSizeFactor(size) {
        switch (size) {
            case "small": return 0.3;
            case "medium": return 0.55;
            case "large": return 0.85;
            case "xlarge": return 1.0;
            default: return 0.5;
        }
    }

    function computeLocalSinglePanelEffect(panel, sizeFactor, panelDist, incidentLight) {
        const solidAngle = sizeFactor / (panelDist * panelDist);
        switch (panel.type) {
            case "bounce_white": {
                const bounced = Math.min(incidentLight * 0.60 * sizeFactor / (panelDist * panelDist), incidentLight * 0.60);
                return { effectIntensity: bounced, softnessModifier: 0.85 + sizeFactor * 0.05, colorTempShift: 0 };
            }
            case "bounce_silver": {
                const bounced = Math.min(incidentLight * 0.85 * sizeFactor / (panelDist * panelDist), incidentLight * 0.85);
                return { effectIntensity: bounced, softnessModifier: 0.85 + sizeFactor * 0.05, colorTempShift: 0 };
            }
            case "bounce_gold": {
                const bounced = Math.min(incidentLight * 0.75 * sizeFactor / (panelDist * panelDist), incidentLight * 0.75);
                return { effectIntensity: bounced, softnessModifier: 0.85 + sizeFactor * 0.05, colorTempShift: 500 };
            }
            case "negative_fill": {
                const absorption = Math.min(incidentLight * 0.3 * solidAngle, incidentLight * 0.25);
                return { effectIntensity: -absorption, softnessModifier: 0, colorTempShift: 0 };
            }
            case "flag": {
                const blocked = Math.min(incidentLight * 0.15 * solidAngle, incidentLight * 0.15);
                return { effectIntensity: -blocked, softnessModifier: 0, colorTempShift: 0 };
            }
            case "diffusion_scrim": {
                const reduction = Math.min(incidentLight * 0.35 * sizeFactor, incidentLight * 0.5);
                return { effectIntensity: -reduction, softnessModifier: 0.95, colorTempShift: 0 };
            }
            default:
                return { effectIntensity: 0, softnessModifier: 0, colorTempShift: 0 };
        }
    }

    function getModifierSoftness(mod) {
        const s = {
            none: 0.1, honeycomb_grid: 0.15, snoot: 0.05, barn_doors: 0.1,
            reflector: 0.3, beauty_dish: 0.5, umbrella: 0.65, softbox: 0.75,
            stripbox: 0.7, octabox: 0.85, diffusion_panel: 0.9, parabolic: 0.6
        };
        return s[mod] || 0.5;
    }

    function applyCSSFilters(f) {
        const subject = document.getElementById("subjectContainer");
        const shadowLayer = document.getElementById("shadowLayer");
        const highlightLayer = document.getElementById("highlightLayer");

        const brightness = f.brightness || 1;
        const contrast = f.contrast || 1;
        const hueRotate = f.hue_rotate || 0;

        subject.style.filter = `brightness(${brightness}) contrast(${contrast}) hue-rotate(${hueRotate}deg)`;

        if (f.shadow_gradient) {
            shadowLayer.style.background = f.shadow_gradient;
        }
        if (f.highlight_pos) {
            highlightLayer.style.background = f.highlight_pos;
        }
    }

    // ---- Utility ----

    function bindRange(inputId, labelId, fn) {
        const input = document.getElementById(inputId);
        input.addEventListener("input", () => {
            document.getElementById(labelId).textContent = fn(input.value);
            updatePreview();
        });
    }

    function bindSelect(id, fn) {
        document.getElementById(id).addEventListener("change", (e) => {
            fn(e.target.value);
            updatePreview();
        });
    }

    function bindCardRange(card, inputSel, labelSel, fn) {
        const input = card.querySelector(inputSel);
        const label = card.querySelector(labelSel);
        input.addEventListener("input", () => {
            label.textContent = fn(input.value);
            updatePreview();
        });
    }

    function refresh() {
        renderDiagram();
        updatePreview();
    }

    // ---- Start ----
    document.addEventListener("DOMContentLoaded", init);
})();
