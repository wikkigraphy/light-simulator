(function () {
    "use strict";

    function init() {
        bindTabs();
        loadFlashGuides();
        loadModifierGuides();
        loadLensGuides();
        loadPatternPresets();
    }

    function bindTabs() {
        document.querySelectorAll(".cs-tab").forEach(tab => {
            tab.addEventListener("click", () => {
                document.querySelectorAll(".cs-tab").forEach(t => t.classList.remove("active"));
                document.querySelectorAll(".cs-content").forEach(c => c.classList.remove("active"));
                tab.classList.add("active");
                document.getElementById(`cstab-${tab.dataset.cstab}`).classList.add("active");
            });
        });
    }

    async function loadFlashGuides() {
        const data = await fetchJSON("/api/guides/flash");
        const container = document.getElementById("flashGuides");
        container.innerHTML = data.map(g => `
            <div class="guide-card">
                <h3>${g.title}</h3>
                <p class="guide-card__desc">${g.description}</p>
                <dl class="guide-card__meta">
                    <dt>Power Range</dt><dd>${g.power_range}</dd>
                    <dt>Recycle Time</dt><dd>${g.recycle_time}</dd>
                </dl>
                <div><strong style="font-size:0.8rem;color:#999">Best For:</strong></div>
                <ul class="guide-card__list">
                    ${g.best_for.map(b => `<li>${b}</li>`).join('')}
                </ul>
                <details class="guide-card__tips">
                    <summary>Pro Tips (${g.tips.length})</summary>
                    <ul>${g.tips.map(t => `<li>${t}</li>`).join('')}</ul>
                </details>
            </div>
        `).join('');
    }

    async function loadModifierGuides() {
        const data = await fetchJSON("/api/guides/modifiers");
        const container = document.getElementById("modifierGuides");
        container.innerHTML = data.map(g => `
            <div class="guide-card">
                <h3>${g.name}</h3>
                <dl class="guide-card__meta">
                    <dt>Size Range</dt><dd>${g.size_range}</dd>
                    <dt>Softness</dt><dd>${g.softness}</dd>
                    <dt>Spill Control</dt><dd>${g.spill_control}</dd>
                    <dt>Catchlight</dt><dd>${g.catchlight_shape}</dd>
                </dl>
                <div><strong style="font-size:0.8rem;color:#999">Best For:</strong></div>
                <ul class="guide-card__list">
                    ${g.best_for.map(b => `<li>${b}</li>`).join('')}
                </ul>
                <details class="guide-card__tips">
                    <summary>Pro Tips (${g.pro_tips.length})</summary>
                    <ul>${g.pro_tips.map(t => `<li>${t}</li>`).join('')}</ul>
                </details>
            </div>
        `).join('');
    }

    async function loadLensGuides() {
        const data = await fetchJSON("/api/guides/lenses");
        const container = document.getElementById("lensGuides");
        container.innerHTML = data.map(g => `
            <div class="guide-card">
                <h3>${g.focal_length} — ${g.type}</h3>
                <dl class="guide-card__meta">
                    <dt>DOF Notes</dt><dd>${g.dof_notes}</dd>
                    <dt>Distortion</dt><dd>${g.distortion}</dd>
                </dl>
                <div><strong style="font-size:0.8rem;color:#999">Best For:</strong></div>
                <ul class="guide-card__list">
                    ${g.best_for.map(b => `<li>${b}</li>`).join('')}
                </ul>
            </div>
        `).join('');
    }

    async function loadPatternPresets() {
        const categories = await fetchJSON("/api/presets");
        const container = document.getElementById("patternPresets");
        let html = "";
        for (const [cat, presets] of Object.entries(categories)) {
            html += `<div class="preset-grid__category"><h2 class="preset-grid__heading">${cat.charAt(0).toUpperCase() + cat.slice(1)}</h2></div>`;
            for (const p of presets) {
                const equipHtml = (p.equipment && p.equipment.length > 0)
                    ? `<div class="preset-card__equipment">
                        <h4 class="preset-card__equipment-title">Equipment List</h4>
                        <table class="preset-card__equipment-table">
                            <thead><tr><th>Role</th><th>Device</th><th>Modifier</th><th>Power</th><th>Placement</th><th>Recommended</th></tr></thead>
                            <tbody>${p.equipment.map(e => `<tr>
                                <td>${e.role}</td><td>${e.device}</td><td>${e.modifier}</td>
                                <td>${e.power}</td><td>${e.placement}</td><td>${e.recommended}</td>
                            </tr>`).join('')}</tbody>
                        </table>
                    </div>`
                    : '';
                html += `
                    <div class="preset-card" data-preset-id="${p.id}" role="button" tabindex="0">
                        <h3>${p.name}</h3>
                        <p>${p.description}</p>
                        <div class="preset-card__lights">
                            ${p.scene.lights.map(l =>
                                `<span class="preset-card__light-tag">${l.name} (${l.role})</span>`
                            ).join('')}
                        </div>
                        ${equipHtml}
                        <div class="preset-card__action">Open in Simulator →</div>
                    </div>
                `;
            }
        }
        container.innerHTML = html;

        container.querySelectorAll(".preset-card[data-preset-id]").forEach(card => {
            card.addEventListener("click", () => {
                window.location = `/simulator?preset=${card.dataset.presetId}`;
            });
            card.addEventListener("keydown", (e) => {
                if (e.key === "Enter" || e.key === " ") {
                    e.preventDefault();
                    window.location = `/simulator?preset=${card.dataset.presetId}`;
                }
            });
        });
    }

    async function fetchJSON(url) {
        try {
            const resp = await fetch(url);
            return await resp.json();
        } catch (e) {
            console.error(`Failed to fetch ${url}:`, e);
            return [];
        }
    }

    document.addEventListener("DOMContentLoaded", init);
})();
