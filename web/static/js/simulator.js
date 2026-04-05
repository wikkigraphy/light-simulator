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

    const LIGHT_TYPE_LABELS = {
        speedlight: "Speedlight", strobe: "Studio Strobe",
        continuous: "Continuous", led_panel: "LED Panel",
        ring_light: "Ring Light", natural: "Natural"
    };

    let scene = {
        id: "custom", name: "Custom Setup", mode: "portrait",
        lights: [], backdrop: "#1a1a1a", ambient: 0.1, notes: "",
        camera: {
            focal_length: 85, aperture: 2.8, shutter_speed: "1/200",
            iso: 100, white_balance: 5500, sensor_size: "full_frame",
            angle_x: 0, angle_y: 0, distance: 2.5
        }
    };

    const CUSTOM_PRESETS_KEY = "light_sim_custom_presets";

    let lightIdCounter = 0;
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
        scene.camera = { ...preset.scene.camera };
        scene.backdrop = preset.scene.backdrop;
        scene.ambient = preset.scene.ambient;

        document.getElementById("shootMode").value = scene.mode;
        syncCameraUI();
        syncSceneUI();
        renderLightsList();
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

        document.getElementById("uploadBtn").addEventListener("click", () => {
            document.getElementById("photoInput").click();
        });

        document.getElementById("photoInput").addEventListener("change", handlePhotoUpload);
        document.getElementById("analyzeBtn").addEventListener("click", analyzeScene);

        document.getElementById("shootMode").addEventListener("change", (e) => {
            scene.mode = e.target.value;
        });
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
                    subjectImg.src = "/static/images/default-subject.png";
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

    function updateLightXZ(light) {
        const rad = light.position.angle * Math.PI / 180;
        light.position.x = Math.sin(rad) * light.position.distance;
        light.position.z = Math.cos(rad) * light.position.distance;
    }

    // ---- Diagram Rendering ----

    function renderDiagram() {
        const group = document.getElementById("lightsGroup");
        group.innerHTML = "";

        for (const light of scene.lights) {
            if (!light.enabled) continue;
            const color = ROLE_COLORS[light.role] || "#888";
            const scale = 75; // 1m = 75px in SVG
            const svgX = light.position.x * scale;
            const svgY = -light.position.z * scale; // SVG Y is inverted

            const g = document.createElementNS("http://www.w3.org/2000/svg", "g");
            g.setAttribute("transform", `translate(${svgX}, ${svgY})`);
            g.setAttribute("data-light-id", light.id);
            g.style.cursor = "grab";

            // Light body
            const circle = document.createElementNS("http://www.w3.org/2000/svg", "circle");
            circle.setAttribute("r", "12");
            circle.setAttribute("fill", color);
            circle.setAttribute("fill-opacity", "0.8");
            circle.setAttribute("stroke", color);
            circle.setAttribute("stroke-width", "1.5");
            g.appendChild(circle);

            // Beam cone
            const beamSpread = getBeamSpread(light);
            const coneLen = 40;
            const halfAngle = (beamSpread / 2) * Math.PI / 180;
            const aimAngle = Math.atan2(-svgY, -svgX); // aim toward subject
            const x1 = Math.cos(aimAngle - halfAngle) * coneLen;
            const y1 = Math.sin(aimAngle - halfAngle) * coneLen;
            const x2 = Math.cos(aimAngle + halfAngle) * coneLen;
            const y2 = Math.sin(aimAngle + halfAngle) * coneLen;

            const cone = document.createElementNS("http://www.w3.org/2000/svg", "path");
            cone.setAttribute("d", `M0,0 L${x1},${y1} L${x2},${y2} Z`);
            cone.setAttribute("fill", color);
            cone.setAttribute("fill-opacity", "0.15");
            cone.setAttribute("stroke", color);
            cone.setAttribute("stroke-opacity", "0.3");
            cone.setAttribute("stroke-width", "0.5");
            g.appendChild(cone);

            // Label
            const text = document.createElementNS("http://www.w3.org/2000/svg", "text");
            text.setAttribute("y", "-16");
            text.setAttribute("text-anchor", "middle");
            text.setAttribute("fill", color);
            text.setAttribute("font-size", "8");
            text.textContent = light.name;
            g.appendChild(text);

            // Line to subject
            const line = document.createElementNS("http://www.w3.org/2000/svg", "line");
            line.setAttribute("x1", "0");
            line.setAttribute("y1", "0");
            line.setAttribute("x2", -svgX);
            line.setAttribute("y2", -svgY);
            line.setAttribute("stroke", color);
            line.setAttribute("stroke-opacity", "0.2");
            line.setAttribute("stroke-width", "0.5");
            line.setAttribute("stroke-dasharray", "3");
            g.appendChild(line);

            group.appendChild(g);
        }

        updateCameraPosition();
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

        svg.addEventListener("mousedown", (e) => {
            const lightG = e.target.closest("[data-light-id]");
            if (!lightG) return;
            const lightId = lightG.getAttribute("data-light-id");
            const light = scene.lights.find(l => l.id === lightId);
            if (!light) return;

            dragState = { light, startX: e.clientX, startY: e.clientY };
            lightG.style.cursor = "grabbing";
            e.preventDefault();
        });

        svg.addEventListener("mousemove", (e) => {
            if (!dragState) return;
            const svgRect = svg.getBoundingClientRect();
            const svgW = svgRect.width;
            const svgH = svgRect.height;

            // Convert mouse position to SVG coordinates
            const mx = ((e.clientX - svgRect.left) / svgW) * 600 - 300;
            const my = ((e.clientY - svgRect.top) / svgH) * 600 - 300;

            const scale = 75;
            dragState.light.position.x = mx / scale;
            dragState.light.position.z = -my / scale;
            dragState.light.position.distance = Math.sqrt(
                dragState.light.position.x ** 2 + dragState.light.position.z ** 2
            );
            dragState.light.position.angle = Math.atan2(
                dragState.light.position.x, dragState.light.position.z
            ) * 180 / Math.PI;
            if (dragState.light.position.angle < 0) {
                dragState.light.position.angle += 360;
            }

            renderDiagram();
            renderLightsList();
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

        for (const light of scene.lights) {
            if (!light.enabled) continue;
            enabledCount++;

            const dist = Math.max(light.position.distance, 0.1);
            const intensity = light.power / (dist * dist);
            totalIntensity += intensity;
            weightedTemp += light.color_temp * intensity;

            const softness = getModifierSoftness(light.modifier);
            avgSoftness += softness;

            if (light.role === "key") {
                keyIntensity = intensity;
                keyDir = light.position.angle;
            }
            if (light.role === "fill") fillIntensity = intensity;
        }

        if (enabledCount > 0) avgSoftness /= enabledCount;

        const filters = { brightness: 1.0, contrast: 1.0, hue_rotate: 0, warmth_shift: 0 };

        if (totalIntensity > 0) {
            const avgTemp = weightedTemp / totalIntensity;
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
        }

        // Shadow gradient
        const shadowAngle = (keyDir + 180) % 360;
        let shadowOpacity = 0.15;
        if (avgSoftness < 0.3) shadowOpacity = 0.6;
        else if (avgSoftness < 0.6) shadowOpacity = 0.35;

        filters.shadow_gradient = `linear-gradient(${shadowAngle}deg, rgba(0,0,0,${shadowOpacity}) 0%, rgba(0,0,0,0) 60%)`;

        // Highlight position
        const hlX = 50 + Math.sin(keyDir * Math.PI / 180) * 30;
        const hlY = 50 - Math.cos(keyDir * Math.PI / 180) * 30;
        filters.highlight_pos = `radial-gradient(circle at ${hlX}% ${hlY}%, rgba(255,255,255,0.15) 0%, rgba(255,255,255,0) 50%)`;

        // Info
        const evDisplay = document.getElementById("evDisplay");
        const ratioDisplay = document.getElementById("ratioDisplay");
        const shadowDisplay = document.getElementById("shadowDisplay");
        const ratio = fillIntensity > 0 ? (keyIntensity / fillIntensity).toFixed(1) : "N/A";
        const shadowQ = avgSoftness > 0.6 ? "soft" : avgSoftness > 0.3 ? "medium" : "hard";

        evDisplay.textContent = `EV: ${(Math.log2(totalIntensity + 1) + 8).toFixed(1)}`;
        ratioDisplay.textContent = `Ratio: ${ratio}${ratio !== "N/A" ? ":1" : ""}`;
        shadowDisplay.textContent = `Shadows: ${shadowQ}`;

        return filters;
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
