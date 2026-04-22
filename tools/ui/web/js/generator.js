(() => {
  // ── Elements ──────────────────────────────────────────────────────────────
  const skillCards = document.querySelectorAll(".skill-card");
  const formService = document.getElementById("form-service");
  const formDomain = document.getElementById("form-domain");
  const outputEl = document.getElementById("gen-output");
  const codeEl = document.getElementById("gen-code");
  const copyBtn = document.getElementById("gen-copy");

  const addMemberBtn = document.getElementById("add-member");
  const membersList = document.getElementById("members-list");
  const membersEmpty = document.getElementById("members-empty");
  const addRepoBtn = document.getElementById("add-repo");
  const reposList = document.getElementById("repos-list");
  const reposEmpty = document.getElementById("repos-empty");

  let currentSkill = "index-service";

  // ── Skill selector ────────────────────────────────────────────────────────
  skillCards.forEach((card) => {
    card.addEventListener("click", () => {
      currentSkill = card.dataset.skill;
      skillCards.forEach((c) => c.classList.toggle("active", c === card));
      formService.classList.toggle("active", currentSkill === "index-service");
      formDomain.classList.toggle("active", currentSkill === "index-domain");
      hideOutput();
    });
  });

  // ── Dynamic rows ──────────────────────────────────────────────────────────
  function syncEmpty(list, emptyEl) {
    emptyEl.hidden = list.children.length > 0;
  }

  function removeBtn(onRemove) {
    const btn = document.createElement("button");
    btn.type = "button";
    btn.className = "dynamic-remove";
    btn.title = "Remove";
    btn.textContent = "×";
    btn.addEventListener("click", onRemove);
    return btn;
  }

  function addMember() {
    const item = document.createElement("div");
    item.className = "dynamic-item";

    const row = document.createElement("div");
    row.className = "field-row";
    row.innerHTML = `
      <div class="field" style="flex:2">
        <input type="text" name="memberName" placeholder="Full Name" />
      </div>
      <div class="field" style="flex:1">
        <select name="memberRole">
          <option value="tech_lead">tech_lead</option>
          <option value="senior_engineer">senior_engineer</option>
          <option value="engineer">engineer</option>
          <option value="product_manager">product_manager</option>
          <option value="data_engineer">data_engineer</option>
          <option value="sre">sre</option>
        </select>
      </div>
      <div class="field" style="flex:1">
        <input type="text" name="memberGithub" placeholder="github_handle" />
      </div>
      <div class="field" style="flex:1">
        <input type="text" name="memberSlack" placeholder="@slack_handle" />
      </div>
    `;

    item.appendChild(row);
    item.appendChild(
      removeBtn(() => {
        item.remove();
        syncEmpty(membersList, membersEmpty);
      }),
    );

    membersList.appendChild(item);
    membersEmpty.hidden = true;
  }

  function addRepo() {
    const item = document.createElement("div");
    item.className = "dynamic-item";

    const row = document.createElement("div");
    row.className = "field-row";
    row.innerHTML = `
      <div class="field" style="flex:3">
        <input type="text" name="repoUrl" placeholder="https://github.com/org/repo" />
      </div>
      <div class="field" style="flex:1">
        <select name="repoOwnership">
          <option value="owner">owner</option>
          <option value="internal">internal</option>
          <option value="partner">partner</option>
        </select>
      </div>
    `;

    item.appendChild(row);
    item.appendChild(
      removeBtn(() => {
        item.remove();
        syncEmpty(reposList, reposEmpty);
      }),
    );

    reposList.appendChild(item);
    reposEmpty.hidden = true;
  }

  addMemberBtn.addEventListener("click", addMember);
  addRepoBtn.addEventListener("click", addRepo);

  // ── Prompt builders ───────────────────────────────────────────────────────
  function buildServicePrompt(fd) {
    const service = fd.get("service")?.trim();
    const domain = fd.get("domain")?.trim();
    const description = fd.get("description")?.trim();
    const ownership = fd.get("ownershipType");
    const criticality = fd.get("criticality");
    const readme = fd.get("readme")?.trim();
    const aiContext = fd.get("aiContext")?.trim();

    const lines = [
      `/index-service`,
      ``,
      `Use the answers below — skip the questionnaire and go directly to repository analysis:`,
      ``,
      `1. Service name: ${service}`,
      `2. Description: ${description}`,
      `3. Ownership type: ${ownership}`,
      `4. Domain: ${domain}`,
      `5. Criticality: ${criticality}`,
    ];

    if (readme || aiContext) {
      lines.push(`6. Docs:`);
      if (readme) lines.push(`   - README: ${readme}`);
      if (aiContext) lines.push(`   - AI_CONTEXT: ${aiContext}`);
    } else {
      lines.push(`6. Docs: find them automatically in the repository.`);
    }

    return lines.join("\n");
  }

  function buildDomainPrompt(fd) {
    const domainName = fd.get("domainName")?.trim();
    const domainDesc = fd.get("domainDesc")?.trim();
    const businessArea = fd.get("businessArea")?.trim();
    const slackChannel = fd.get("slackChannel")?.trim();
    const outputName = fd.get("outputName")?.trim();
    const teamName = fd.get("teamName")?.trim();
    const teamMission = fd.get("teamMission")?.trim();

    const members = Array.from(membersList.querySelectorAll(".dynamic-item"))
      .map((item) => ({
        name: item.querySelector('[name="memberName"]').value.trim(),
        role: item.querySelector('[name="memberRole"]').value,
        github: item.querySelector('[name="memberGithub"]').value.trim(),
        slack: item.querySelector('[name="memberSlack"]').value.trim(),
      }))
      .filter((m) => m.name);

    const repos = Array.from(reposList.querySelectorAll(".dynamic-item"))
      .map((item) => ({
        url: item.querySelector('[name="repoUrl"]').value.trim(),
        ownership: item.querySelector('[name="repoOwnership"]').value,
      }))
      .filter((r) => r.url);

    const lines = [
      `/index-domain`,
      ``,
      `Use the answers below — skip the questionnaire and go directly to repository scraping:`,
      ``,
      `**Domain**`,
      `1. Domain name: ${domainName}`,
      `2. Description: ${domainDesc}`,
      `3. Business area: ${businessArea || "TODO"}`,
      `4. Slack channel: ${slackChannel || "TODO"}`,
      ``,
      `**Team**`,
      `5. Team name: ${teamName || "TODO"}`,
      `6. Mission: ${teamMission || "TODO"}`,
    ];

    if (members.length > 0) {
      lines.push(`7. Members:`);
      members.forEach((m) =>
        lines.push(
          `   ${m.name} | ${m.role} | ${m.github || "TODO"} | @${m.slack || "TODO"}`,
        ),
      );
    } else {
      lines.push(`7. Members: TODO`);
    }

    lines.push(``, `**Repositories**`);

    if (repos.length > 0) {
      lines.push(`8. Repositories:`);
      repos.forEach((r) => lines.push(`   ${r.url} | ${r.ownership}`));
    } else {
      lines.push(`8. Repositories: TODO`);
    }

    lines.push(``, `**Output**`, `9. Output file name: ${outputName}`);

    return lines.join("\n");
  }

  // ── Output helpers ────────────────────────────────────────────────────────
  function showOutput(prompt) {
    codeEl.textContent = prompt;
    outputEl.hidden = false;
    outputEl.scrollIntoView({ behavior: "smooth", block: "start" });
  }

  function hideOutput() {
    outputEl.hidden = true;
    codeEl.textContent = "";
  }

  // ── Submit ────────────────────────────────────────────────────────────────
  function handleSubmit(e) {
    e.preventDefault();
    const fd = new FormData(e.target);

    if (currentSkill === "index-service") {
      if (
        !fd.get("service")?.trim() ||
        !fd.get("domain")?.trim() ||
        !fd.get("description")?.trim()
      ) {
        e.target.querySelector("[required]")?.focus();
        return;
      }
      showOutput(buildServicePrompt(fd));
    } else {
      if (
        !fd.get("domainName")?.trim() ||
        !fd.get("domainDesc")?.trim() ||
        !fd.get("outputName")?.trim()
      ) {
        e.target.querySelector("[required]")?.focus();
        return;
      }
      showOutput(buildDomainPrompt(fd));
    }
  }

  // Reset only clears the active form's dynamic lists
  function handleServiceReset() {
    hideOutput();
  }

  function handleDomainReset() {
    membersList.innerHTML = "";
    reposList.innerHTML = "";
    membersEmpty.hidden = false;
    reposEmpty.hidden = false;
    hideOutput();
  }

  formService.addEventListener("submit", handleSubmit);
  formDomain.addEventListener("submit", handleSubmit);
  formService.addEventListener("reset", handleServiceReset);
  formDomain.addEventListener("reset", handleDomainReset);

  // ── Copy ──────────────────────────────────────────────────────────────────
  copyBtn.addEventListener("click", async () => {
    try {
      await navigator.clipboard.writeText(codeEl.textContent);
      const orig = copyBtn.textContent;
      copyBtn.textContent = "Copied!";
      setTimeout(() => {
        copyBtn.textContent = orig;
      }, 1500);
    } catch {
      copyBtn.textContent = "Failed";
      setTimeout(() => {
        copyBtn.textContent = "Copy";
      }, 1500);
    }
  });
})();
