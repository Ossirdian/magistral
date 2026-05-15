(function () {
  const app = document.getElementById("app");

  const state = {
    session: readJSON("magistral.session"),
    me: null,
    view: "login",
    busy: false,
    toast: null,
    schools: [],
    branches: [],
    users: [],
    cars: [],
    instructors: [],
    availability: [],
    selectedSchoolId: null,
    selectedResourceType: "cars",
    selectedResourceId: null,
    drafts: {
      schools: {},
      branches: {},
      users: {},
      cars: {},
      instructors: {},
      availability: [],
    },
  };

  const weekDays = ["Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"];

  start().catch(function (error) {
    writeSession(null);
    state.me = null;
    state.view = "login";
    showError(error);
  });

  async function start() {
    render();
    if (!state.session) return;
    await loadMe();
    await routeAfterLogin();
  }

  function readJSON(key) {
    try {
      return JSON.parse(localStorage.getItem(key) || "null");
    } catch (_error) {
      return null;
    }
  }

  function writeSession(session) {
    state.session = session;
    if (session) {
      localStorage.setItem("magistral.session", JSON.stringify(session));
    } else {
      localStorage.removeItem("magistral.session");
    }
  }

  async function api(base, path, options, retry) {
    const requestOptions = options || {};
    const headers = Object.assign({ "Content-Type": "application/json" }, requestOptions.headers || {});

    if (state.session && state.session.accessToken) {
      headers.Authorization = "Bearer " + state.session.accessToken;
    }

    const response = await fetch(base + path, Object.assign({}, requestOptions, { headers: headers }));
    if (response.status === 401 && retry !== false && state.session && state.session.refreshToken) {
      await refresh();
      return api(base, path, options, false);
    }

    const payload = response.status === 204 ? null : await response.json().catch(function () { return null; });
    if (!response.ok) {
      throw new Error(payload && payload.error && payload.error.message ? payload.error.message : "HTTP " + response.status);
    }

    return payload;
  }

  const iam = function (path, options) { return api("/api/iam", path, options); };
  const saas = function (path, options) { return api("/api/saas", path, options); };
  const resources = function (path, options) { return api("/api/resources", path, options); };

  async function refresh() {
    const response = await fetch("/api/iam/auth/refresh", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: state.session.refreshToken }),
    });
    if (!response.ok) throw new Error("refresh failed");
    const payload = await response.json();
    writeSession(normalizeSession(payload.data));
  }

  function normalizeSession(data) {
    return {
      accessToken: data.access_token,
      refreshToken: data.refresh_token,
      expiresAt: data.expires_at,
      identity: data.identity,
    };
  }

  async function loadMe() {
    const payload = await iam("/auth/me");
    state.me = payload.data;
  }

  async function routeAfterLogin() {
    if (state.me.role === "super_admin") {
      state.view = "super-admin";
      await loadSaas();
    } else {
      state.view = "resources";
      await loadResources();
    }
    render();
  }

  async function loadSaas() {
    const schoolsPayload = await saas("/schools?limit=100&offset=0");
    state.schools = schoolsPayload.data || [];

    if (!state.selectedSchoolId || !findById(state.schools, state.selectedSchoolId)) {
      state.selectedSchoolId = state.schools[0] ? state.schools[0].id : null;
    }

    if (!state.selectedSchoolId) {
      state.branches = [];
      state.users = [];
    } else {
      const payloads = await Promise.all([
        saas("/branches?school_id=" + encodeURIComponent(state.selectedSchoolId) + "&limit=100&offset=0"),
        iam("/admin/users?school_id=" + encodeURIComponent(state.selectedSchoolId) + "&limit=100&offset=0"),
      ]);
      state.branches = payloads[0].data || [];
      state.users = payloads[1].data && payloads[1].data.users ? payloads[1].data.users : [];
    }

    state.drafts.schools = buildDraftMap(state.schools, schoolDraft);
    state.drafts.branches = buildDraftMap(state.branches, branchDraft);
    state.drafts.users = buildDraftMap(state.users, userDraft);
  }

  async function loadResources() {
    const requests = [
      resources("/resources/cars?limit=100&offset=0"),
      resources("/resources/instructors?limit=100&offset=0"),
    ];

    if (state.me && state.me.school_id) {
      requests.push(saas("/branches?school_id=" + encodeURIComponent(state.me.school_id) + "&limit=100&offset=0"));
    }

    const payloads = await Promise.all(requests);
    state.cars = payloads[0].data || [];
    state.instructors = payloads[1].data || [];
    state.branches = payloads[2] && payloads[2].data ? payloads[2].data : state.branches;

    const collection = currentResourceCollection();
    if (!state.selectedResourceId || !findById(collection, state.selectedResourceId)) {
      state.selectedResourceId = collection[0] ? collection[0].id : null;
    }

    state.drafts.cars = buildDraftMap(state.cars, carDraft);
    state.drafts.instructors = buildDraftMap(state.instructors, instructorDraft);
    await loadAvailability();
  }

  async function loadAvailability() {
    if (!state.selectedResourceId) {
      state.availability = [];
      state.drafts.availability = [];
      return;
    }

    const path = state.selectedResourceType === "cars"
      ? "/resources/cars/" + state.selectedResourceId + "/availability"
      : "/resources/instructors/" + state.selectedResourceId + "/availability";

    const payload = await resources(path);
    state.availability = payload.data || [];
    state.drafts.availability = state.availability.map(availabilityDraft);
  }

  function render() {
    if (!state.me) {
      app.innerHTML = renderLogin();
    } else if (state.view === "super-admin") {
      app.innerHTML = renderSuperAdmin();
    } else {
      app.innerHTML = renderResources();
    }
    bind();
    refreshSaveDock();
  }

  function bind() {
    app.onclick = onClick;
    app.onchange = onChange;
    app.onsubmit = onSubmit;
  }

  function renderLogin() {
    return [
      '<div class="login-layout">',
      '  <section class="login-hero">',
      '    <div class="hero-badge">MAGISTRAL CRM</div>',
      '    <h1>Операционная панель для школ, ресурсов и администрирования</h1>',
      '    <p>Белый и ярко-зеленый интерфейс в стиле обычной CRM: каталоги, рабочие списки, карточки и сохранение изменений без прыжков по экрану.</p>',
      '    <div class="hero-metrics">',
      metric("SaaS", "Школы и филиалы"),
      metric("IAM", "Пользователи и роли"),
      metric("Resources", "Автомобили и инструкторы"),
      "    </div>",
      "  </section>",
      '  <section class="login-panel">',
      '    <div class="panel-head compact"><div><span class="eyebrow">Авторизация</span><h2>Вход в систему</h2></div></div>',
      renderToast(),
      '    <form id="login-form" class="form-stack">',
      '      <label>Email<input name="email" type="email" placeholder="superadmin@example.com" required></label>',
      '      <label>Пароль<input name="password" type="password" placeholder="Passw0rd!" required></label>',
      '      <button class="primary-button big" type="submit">' + (state.busy ? "Входим..." : "Войти") + "</button>",
      "    </form>",
      '    <div class="hint-box"><strong>Стартовый доступ</strong><span>superadmin@example.com / Passw0rd!</span></div>',
      "  </section>",
      "</div>",
    ].join("");
  }

  function renderSuperAdmin() {
    const selectedSchool = findById(state.schools, state.selectedSchoolId);
    const schoolRows = state.schools.map(renderSchoolCard).join("");
    const branchRows = state.branches.map(renderBranchRow).join("");
    const userRows = state.users.map(renderUserRow).join("");

    return shell(
      {
        title: "CRM для super admin",
        subtitle: "Управление школами, филиалами и доступами в одном интерфейсе.",
        nav: [
          navButton("Панель SaaS", true),
          navButton("Обновить данные", false, "reload-saas"),
        ].join(""),
      },
      [
        '<section class="hero-strip">',
        kpi("Школы", String(state.schools.length), "Каталог организаций"),
        kpi("Филиалы", String(state.branches.length), "По выбранной школе"),
        kpi("Пользователи", String(state.users.length), "IAM записи"),
        kpi("Черновики", String(saveSummary().count), "Ожидают сохранения"),
        "</section>",
        '<section class="workspace-grid super-admin-grid">',
        '  <div class="panel tall">',
        '    <div class="panel-head"><div><span class="eyebrow">Каталог школ</span><h3>Список школ</h3></div></div>',
        '    <div class="entity-list">' + (schoolRows || '<div class="empty-state">Школы пока не созданы.</div>') + "</div>",
        "  </div>",
        '  <div class="stack-column">',
        '    <div class="panel">',
        '      <div class="panel-head"><div><span class="eyebrow">Карточка школы</span><h3>Редактор школы</h3></div></div>',
        renderSchoolEditor(selectedSchool),
        "    </div>",
        '    <div class="panel">',
        '      <div class="panel-head"><div><span class="eyebrow">Создание</span><h3>Новая школа</h3></div></div>',
        renderCreateSchoolForms(),
        "    </div>",
        "  </div>",
        "</section>",
        '<section class="workspace-grid two-col">',
        '  <div class="panel">',
        '    <div class="panel-head"><div><span class="eyebrow">Филиалы</span><h3>Редактируемый список</h3></div></div>',
        renderCreateBranchForm(),
        table(["Название", "Адрес", "Статус"], branchRows, "Филиалы появятся после выбора школы."),
        "  </div>",
        '  <div class="panel">',
        '    <div class="panel-head"><div><span class="eyebrow">IAM</span><h3>Редактируемый список пользователей</h3></div></div>',
        renderCreateUserForm(),
        table(["Email", "Роль", "Статус", "Активен"], userRows, "Пользователи появятся после выбора школы."),
        "  </div>",
        "</section>",
        renderSaveDock(),
      ].join("")
    );
  }

  function renderResources() {
    const collection = currentResourceCollection();
    const selected = findById(collection, state.selectedResourceId);
    const rows = collection.map(renderResourceRow).join("");

    return shell(
      {
        title: "CRM ресурсов",
        subtitle: "Автомобили, инструкторы и расписание доступности.",
        nav: [
          navButton("Автомобили", state.selectedResourceType === "cars", "switch-type", "cars"),
          navButton("Инструкторы", state.selectedResourceType === "instructors", "switch-type", "instructors"),
          navButton("Обновить", false, "reload-resources"),
        ].join(""),
      },
      [
        '<section class="hero-strip">',
        kpi("Автомобили", String(state.cars.length), "Рабочий парк"),
        kpi("Инструкторы", String(state.instructors.length), "Команда"),
        kpi("Филиалы", String(state.branches.length), "Справочник"),
        kpi("Изменения", String(saveSummary().count), "Есть несохраненные правки"),
        "</section>",
        '<section class="workspace-grid resources-grid">',
        '  <div class="panel tall">',
        '    <div class="panel-head"><div><span class="eyebrow">' + esc(state.selectedResourceType === "cars" ? "Автопарк" : "Команда") + '</span><h3>Редактируемый список</h3></div></div>',
        renderResourceCreateForm(),
        table(
          state.selectedResourceType === "cars"
            ? ["Госномер", "Филиал", "Марка", "Модель", "Цвет", "Статус"]
            : ["ФИО", "Филиал", "Email", "Телефон", "Статус"],
          rows,
          "Список ресурсов пуст."
        ),
        "  </div>",
        '  <div class="stack-column">',
        '    <div class="panel">',
        '      <div class="panel-head"><div><span class="eyebrow">Карточка</span><h3>Выбранный ресурс</h3></div></div>',
        renderResourceCard(selected),
        "    </div>",
        '    <div class="panel">',
        '      <div class="panel-head"><div><span class="eyebrow">Availability</span><h3>Редактор расписания</h3></div></div>',
        renderAvailabilityEditor(selected),
        "    </div>",
        (state.selectedResourceType === "cars" ? renderAssignForm() : ""),
        "  </div>",
        "</section>",
        renderSaveDock(),
      ].join("")
    );
  }

  function shell(config, body) {
    return [
      '<div class="crm-shell">',
      '  <aside class="sidebar">',
      '    <div class="brand-block"><div class="brand-mark">M</div><div><span class="eyebrow bright">Magistral</span><h1>Control CRM</h1><p>' + esc(readableRole(state.me.role)) + "</p></div></div>",
      '    <div class="nav-stack">' + config.nav + "</div>",
      '    <div class="sidebar-card"><div class="sidebar-label">Профиль</div><strong>' + esc(state.me.email || state.me.user_id || "user") + "</strong><span>" + esc(state.me.role) + "</span>" + (state.me.school_id ? '<span class="muted-light">' + esc(state.me.school_id) + "</span>" : "") + "</div>",
      '    <button class="ghost-danger" data-action="logout">Выйти</button>',
      "  </aside>",
      '  <main class="main-area">',
      '    <header class="page-head"><div><span class="eyebrow">Рабочее место</span><h2>' + esc(config.title) + "</h2><p>" + esc(config.subtitle) + '</p></div><div class="head-badges">' + (state.busy ? '<span class="busy-pill">Выполняется запрос</span>' : "") + '<span class="state-pill">' + esc(new Date().toLocaleDateString("ru-RU")) + "</span></div></header>",
      renderToast(),
      body,
      "  </main>",
      "</div>",
    ].join("");
  }

  function metric(value, label) {
    return '<div class="metric-card"><strong>' + esc(value) + "</strong><span>" + esc(label) + "</span></div>";
  }

  function kpi(label, value, hint) {
    return '<article class="kpi-card"><span class="eyebrow">' + esc(label) + "</span><strong>" + esc(value) + "</strong><p>" + esc(hint) + "</p></article>";
  }

  function navButton(label, active, action, value) {
    return '<button class="nav-button ' + (active ? "active" : "") + '"' + (action ? ' data-action="' + esc(action) + '"' : "") + (value ? ' data-value="' + esc(value) + '"' : "") + ">" + esc(label) + "</button>";
  }

  function renderSchoolCard(item) {
    const draft = state.drafts.schools[item.id];
    return [
      '<button class="entity-card ' + (item.id === state.selectedSchoolId ? "selected" : "") + '" data-action="select-school" data-id="' + esc(item.id) + '">',
      '  <div class="entity-card-head"><strong>' + esc(draft.name) + '</strong><span class="state-pill ' + (isDirty("schools", item.id) ? "accent" : "") + '">' + esc(draft.status) + "</span></div>",
      '  <span class="entity-meta">' + esc(item.id) + "</span>",
      "</button>",
    ].join("");
  }

  function renderSchoolEditor(item) {
    if (!item) return '<div class="empty-state">Выберите школу слева, чтобы редактировать ее карточку.</div>';
    const draft = state.drafts.schools[item.id];
    return [
      '<div class="editor-grid single">',
      '  <label>Название<input data-scope="schools" data-id="' + esc(item.id) + '" data-field="name" value="' + escAttr(draft.name) + '"></label>',
      '  <label>Статус<select data-scope="schools" data-id="' + esc(item.id) + '" data-field="status">' + options(["active", "inactive", "suspended"], draft.status) + "</select></label>",
      '  <div class="record-caption">Создано: ' + esc(formatDateTime(item.created_at)) + " · Обновлено: " + esc(formatDateTime(item.updated_at)) + "</div>",
      "</div>",
    ].join("");
  }

  function renderCreateSchoolForms() {
    return [
      '<div class="stack-column">',
      '  <form id="create-school-form" class="editor-grid single form-block">',
      '    <label>Название<input name="name" required></label>',
      '    <label>Статус<select name="status">' + options(["active", "inactive", "suspended"], "active") + "</select></label>",
      '    <button class="primary-button" type="submit">Создать школу</button>',
      "  </form>",
      '  <form id="bootstrap-school-form" class="editor-grid form-block">',
      '    <label>Название<input name="name" required></label>',
      '    <label>Статус<select name="status">' + options(["active", "inactive", "suspended"], "active") + "</select></label>",
      '    <label>Email первого admin<input name="admin_email" type="email" required></label>',
      '    <label>Пароль admin<input name="password" type="password" required></label>',
      '    <button class="primary-button" type="submit">Bootstrap школа + admin</button>',
      "  </form>",
      "</div>",
    ].join("");
  }

  function renderCreateBranchForm() {
    return '<form id="create-branch-form" class="toolbar-form"><input name="name" placeholder="Новый филиал" ' + (state.selectedSchoolId ? "" : "disabled") + ' required><input name="address" placeholder="Адрес" ' + (state.selectedSchoolId ? "" : "disabled") + '><select name="status" ' + (state.selectedSchoolId ? "" : "disabled") + ">" + options(["active", "inactive"], "active") + '</select><button class="primary-button" type="submit" ' + (state.selectedSchoolId ? "" : "disabled") + ">Добавить</button></form>";
  }

  function renderCreateUserForm() {
    return '<form id="create-user-form" class="toolbar-form wide"><input name="email" type="email" placeholder="user@example.com" ' + (state.selectedSchoolId ? "" : "disabled") + ' required><input name="password" type="password" placeholder="Пароль" ' + (state.selectedSchoolId ? "" : "disabled") + ' required><select name="role" ' + (state.selectedSchoolId ? "" : "disabled") + ">" + options(["admin", "owner", "instructor"], "admin") + '</select><select name="status" ' + (state.selectedSchoolId ? "" : "disabled") + ">" + options(["active", "inactive", "suspended"], "active") + '</select><button class="primary-button" type="submit" ' + (state.selectedSchoolId ? "" : "disabled") + ">Добавить</button></form>";
  }

  function renderBranchRow(item) {
    const draft = state.drafts.branches[item.id];
    return "<tr><td><input data-scope=\"branches\" data-id=\"" + esc(item.id) + '" data-field="name" value="' + escAttr(draft.name) + '"></td><td><input data-scope="branches" data-id="' + esc(item.id) + '" data-field="address" value="' + escAttr(draft.address || "") + '"></td><td><select data-scope="branches" data-id="' + esc(item.id) + '" data-field="status">' + options(["active", "inactive"], draft.status) + "</select></td></tr>";
  }

  function renderUserRow(item) {
    const draft = state.drafts.users[item.id];
    return "<tr><td><input data-scope=\"users\" data-id=\"" + esc(item.id) + '" data-field="email" value="' + escAttr(draft.email) + '"></td><td><select data-scope="users" data-id="' + esc(item.id) + '" data-field="role">' + options(["admin", "owner", "instructor"], draft.role) + '</select></td><td><select data-scope="users" data-id="' + esc(item.id) + '" data-field="status">' + options(["active", "inactive", "suspended"], draft.status) + '</select></td><td><label class="switch"><input type="checkbox" data-scope="users" data-id="' + esc(item.id) + '" data-field="is_active" ' + (draft.is_active ? "checked" : "") + "><span></span></label></td></tr>";
  }

  function renderResourceCreateForm() {
    if (!(state.me.role === "admin" || state.me.role === "owner")) return "";
    if (state.selectedResourceType === "cars") {
      return '<form id="create-car-form" class="toolbar-form wide">' + branchSelect("branch_id", null, true) + '<input name="plate_number" placeholder="Госномер" required><input name="brand" placeholder="Марка"><input name="model" placeholder="Модель"><input name="color" placeholder="Цвет"><button class="primary-button" type="submit">Добавить авто</button></form>';
    }
    return '<form id="create-instructor-form" class="toolbar-form wide">' + branchSelect("branch_id", null, true) + '<input name="full_name" placeholder="ФИО" required><input name="email" type="email" placeholder="Email"><input name="phone" placeholder="Телефон"><button class="primary-button" type="submit">Добавить инструктора</button></form>';
  }

  function renderResourceRow(item) {
    const draft = state.drafts[state.selectedResourceType][item.id];
    const rowClass = item.id === state.selectedResourceId ? ' class="active-row"' : "";

    if (state.selectedResourceType === "cars") {
      return "<tr" + rowClass + ' data-action="select-resource" data-id="' + esc(item.id) + '"><td><input data-scope="cars" data-id="' + esc(item.id) + '" data-field="plate_number" value="' + escAttr(draft.plate_number) + '"></td><td>' + branchSelect(null, draft.branch_id, false, "cars", item.id) + '</td><td><input data-scope="cars" data-id="' + esc(item.id) + '" data-field="brand" value="' + escAttr(draft.brand || "") + '"></td><td><input data-scope="cars" data-id="' + esc(item.id) + '" data-field="model" value="' + escAttr(draft.model || "") + '"></td><td><input data-scope="cars" data-id="' + esc(item.id) + '" data-field="color" value="' + escAttr(draft.color || "") + '"></td><td><select data-scope="cars" data-id="' + esc(item.id) + '" data-field="status">' + options(["active", "maintenance", "inactive"], draft.status) + "</select></td></tr>";
    }

    return "<tr" + rowClass + ' data-action="select-resource" data-id="' + esc(item.id) + '"><td><input data-scope="instructors" data-id="' + esc(item.id) + '" data-field="full_name" value="' + escAttr(draft.full_name) + '"></td><td>' + branchSelect(null, draft.branch_id, false, "instructors", item.id) + '</td><td><input data-scope="instructors" data-id="' + esc(item.id) + '" data-field="email" value="' + escAttr(draft.email || "") + '"></td><td><input data-scope="instructors" data-id="' + esc(item.id) + '" data-field="phone" value="' + escAttr(draft.phone || "") + '"></td><td><select data-scope="instructors" data-id="' + esc(item.id) + '" data-field="status">' + options(["active", "inactive", "on_leave"], draft.status) + "</select></td></tr>";
  }

  function renderResourceCard(item) {
    if (!item) return '<div class="empty-state">Выберите ресурс в таблице слева.</div>';
    return '<div class="detail-grid">' +
      detail("ID", item.id) +
      detail("Филиал", branchName(item.branch_id)) +
      detail("Статус", item.status) +
      detail("Availability", item.availability_status) +
      detail("Обновлено", formatDateTime(item.updated_at)) +
      detail("Создано", formatDateTime(item.created_at)) +
      "</div>";
  }

  function renderAvailabilityEditor(item) {
    if (!item) return '<div class="empty-state">Сначала выберите ресурс.</div>';

    const rows = state.drafts.availability.map(function (window) {
      return '<div class="availability-row"><select data-scope="availability" data-id="' + esc(window.id) + '" data-field="type">' + options(["working_hours", "vacation", "blackout"], window.type) + '</select><input data-scope="availability" data-id="' + esc(window.id) + '" data-field="start_at" type="datetime-local" value="' + escAttr(window.start_at) + '"><input data-scope="availability" data-id="' + esc(window.id) + '" data-field="end_at" type="datetime-local" value="' + escAttr(window.end_at) + '"><input data-scope="availability" data-id="' + esc(window.id) + '" data-field="comment" value="' + escAttr(window.comment || "") + '" placeholder="Комментарий"><button type="button" class="icon-button" data-action="remove-availability" data-id="' + esc(window.id) + '">Удалить</button></div>';
    }).join("");

    return renderCalendarPreview() + '<div class="availability-list">' + (rows || '<div class="empty-state compact">Окон пока нет.</div>') + '</div><button type="button" class="ghost-button" data-action="add-availability">Добавить окно</button>';
  }

  function renderAssignForm() {
    if (!(state.me.role === "admin" || state.me.role === "owner")) return "";
    return '<div class="panel"><div class="panel-head"><div><span class="eyebrow">Связка</span><h3>Назначение авто инструктору</h3></div></div><form id="assign-car-form" class="editor-grid single"><label>Инструктор<select name="instructor_id">' + state.instructors.map(function (item) { return '<option value="' + esc(item.id) + '">' + esc(item.full_name) + "</option>"; }).join("") + '</select></label><label>Effective from<input name="effective_from" type="datetime-local"></label><label>Комментарий<input name="comment"></label><button class="primary-button" type="submit" ' + (state.selectedResourceId ? "" : "disabled") + ">Назначить</button></form></div>";
  }

  function renderCalendarPreview() {
    const grouped = [[], [], [], [], [], [], []];
    state.drafts.availability.forEach(function (item) {
      const index = (new Date(toISO(item.start_at)).getDay() + 6) % 7;
      grouped[index].push(item);
    });

    return '<div class="calendar-grid">' + grouped.map(function (items, index) {
      return '<div class="calendar-day"><h4>' + esc(weekDays[index]) + "</h4>" + (items.length
        ? items.map(function (item) {
            return '<div class="window-chip ' + esc(item.type) + '"><strong>' + esc(item.type) + '</strong><span>' + esc(shortWindow(item.start_at, item.end_at)) + "</span></div>";
          }).join("")
        : '<span class="muted-text">Нет окон</span>') + "</div>";
    }).join("") + "</div>";
  }

  function table(headers, rows, emptyMessage) {
    return '<div class="table-shell"><table class="crm-table"><thead><tr>' + headers.map(function (header) { return "<th>" + esc(header) + "</th>"; }).join("") + '</tr></thead><tbody>' + (rows || '<tr><td colspan="' + headers.length + '"><div class="empty-state compact">' + esc(emptyMessage) + "</div></td></tr>") + "</tbody></table></div>";
  }

  function detail(label, value) {
    return '<div class="detail-card"><span>' + esc(label) + "</span><strong>" + esc(value || "—") + "</strong></div>";
  }

  function renderSaveDock() {
    const summary = saveSummary();
    return '<div id="save-dock" class="save-dock ' + (summary.count ? "visible" : "") + '"><div class="save-dock-copy"><strong>' + (summary.count ? "Есть несохраненные изменения" : "Изменений нет") + "</strong><span>" + esc(summary.label) + '</span></div><div class="save-dock-actions"><button class="ghost-button" data-action="discard-drafts" ' + (summary.count ? "" : "disabled") + '>Отменить</button><button class="primary-button" data-action="save-drafts" ' + (summary.count ? "" : "disabled") + '>' + (state.busy ? "Сохраняем..." : "Сохранить изменения") + "</button></div></div>";
  }

  function renderToast() {
    if (!state.toast) return "";
    return '<div class="toast ' + (state.toast.type === "error" ? "error" : "success") + '"><span>' + esc(state.toast.message) + '</span><button class="icon-link" data-action="clear-toast">Закрыть</button></div>';
  }

  function onClick(event) {
    const target = event.target.closest("[data-action]");
    if (!target) return;

    const action = target.dataset.action;
    if (action === "logout") {
      writeSession(null);
      state.me = null;
      state.view = "login";
      render();
      return;
    }
    if (action === "clear-toast") {
      state.toast = null;
      render();
      return;
    }
    if (action === "reload-saas") return withBusy(async function () { await loadSaas(); render(); });
    if (action === "reload-resources") return withBusy(async function () { await loadResources(); render(); });
    if (action === "select-school") return withBusy(async function () { state.selectedSchoolId = target.dataset.id; await loadSaas(); render(); });
    if (action === "switch-type") return withBusy(async function () { state.selectedResourceType = target.dataset.value; state.selectedResourceId = null; await loadResources(); render(); });
    if (action === "select-resource") return withBusy(async function () { state.selectedResourceId = target.dataset.id; await loadAvailability(); render(); });
    if (action === "add-availability") {
      state.drafts.availability.push({
        id: "draft-" + Date.now(),
        type: "working_hours",
        start_at: nextLocal(9),
        end_at: nextLocal(18),
        comment: "",
      });
      render();
      return;
    }
    if (action === "remove-availability") {
      state.drafts.availability = state.drafts.availability.filter(function (item) { return item.id !== target.dataset.id; });
      render();
      return;
    }
    if (action === "save-drafts") return saveDrafts();
    if (action === "discard-drafts") return discardDrafts();
  }

  function onChange(event) {
    const node = event.target;
    if (!node.dataset || !node.dataset.scope) return;

    const scope = node.dataset.scope;
    const id = node.dataset.id;
    const field = node.dataset.field;
    const value = node.type === "checkbox" ? node.checked : node.value;

    if (scope === "availability") {
      const item = state.drafts.availability.find(function (window) { return window.id === id; });
      if (item) item[field] = value;
    } else if (state.drafts[scope] && state.drafts[scope][id]) {
      state.drafts[scope][id][field] = value === "" && nullableField(field) ? null : value;
    }

    refreshSaveDock();
  }

  function onSubmit(event) {
    event.preventDefault();
    const form = event.target;

    if (form.id === "login-form") return onLogin(form);
    if (form.id === "create-school-form") return onCreateSchool(form);
    if (form.id === "bootstrap-school-form") return onBootstrapSchool(form);
    if (form.id === "create-branch-form") return onCreateBranch(form);
    if (form.id === "create-user-form") return onCreateUser(form);
    if (form.id === "create-car-form") return onCreateCar(form);
    if (form.id === "create-instructor-form") return onCreateInstructor(form);
    if (form.id === "assign-car-form") return onAssignCar(form);
  }

  async function onLogin(form) {
    await withBusy(async function () {
      const data = new FormData(form);
      const payload = await iam("/auth/login", {
        method: "POST",
        body: JSON.stringify({
          email: data.get("email"),
          password: data.get("password"),
        }),
      });
      writeSession(normalizeSession(payload.data));
      await loadMe();
      toast("success", "Сессия открыта.");
      await routeAfterLogin();
    });
  }

  async function onCreateSchool(form) {
    await withBusy(async function () {
      const data = new FormData(form);
      await saas("/schools", {
        method: "POST",
        body: JSON.stringify({
          name: data.get("name"),
          status: data.get("status"),
        }),
      });
      form.reset();
      await loadSaas();
      render();
      toast("success", "Школа создана.");
    });
  }

  async function onBootstrapSchool(form) {
    await withBusy(async function () {
      const data = new FormData(form);
      await saas("/admin/bootstrap-school", {
        method: "POST",
        body: JSON.stringify({
          name: data.get("name"),
          status: data.get("status"),
          admin_email: data.get("admin_email"),
          password: data.get("password"),
        }),
      });
      form.reset();
      await loadSaas();
      render();
      toast("success", "Школа и первый admin созданы.");
    });
  }

  async function onCreateBranch(form) {
    await withBusy(async function () {
      const data = new FormData(form);
      await saas("/branches", {
        method: "POST",
        body: JSON.stringify({
          school_id: state.selectedSchoolId,
          name: data.get("name"),
          address: emptyToNull(data.get("address")),
          status: data.get("status"),
        }),
      });
      form.reset();
      await loadSaas();
      render();
      toast("success", "Филиал создан.");
    });
  }

  async function onCreateUser(form) {
    await withBusy(async function () {
      const data = new FormData(form);
      await iam("/admin/users", {
        method: "POST",
        body: JSON.stringify({
          school_id: state.selectedSchoolId,
          email: data.get("email"),
          password: data.get("password"),
          role: data.get("role"),
          status: data.get("status"),
        }),
      });
      form.reset();
      await loadSaas();
      render();
      toast("success", "Пользователь создан.");
    });
  }

  async function onCreateCar(form) {
    await withBusy(async function () {
      const data = new FormData(form);
      await resources("/resources/cars", {
        method: "POST",
        body: JSON.stringify({
          branch_id: data.get("branch_id"),
          plate_number: data.get("plate_number"),
          brand: emptyToNull(data.get("brand")),
          model: emptyToNull(data.get("model")),
          color: emptyToNull(data.get("color")),
          status: "active",
        }),
      });
      form.reset();
      await loadResources();
      render();
      toast("success", "Автомобиль создан.");
    });
  }

  async function onCreateInstructor(form) {
    await withBusy(async function () {
      const data = new FormData(form);
      await resources("/resources/instructors", {
        method: "POST",
        body: JSON.stringify({
          branch_id: data.get("branch_id"),
          full_name: data.get("full_name"),
          email: emptyToNull(data.get("email")),
          phone: emptyToNull(data.get("phone")),
          status: "active",
        }),
      });
      form.reset();
      await loadResources();
      render();
      toast("success", "Инструктор создан.");
    });
  }

  async function onAssignCar(form) {
    await withBusy(async function () {
      const data = new FormData(form);
      await resources("/resources/cars/" + state.selectedResourceId + "/assign", {
        method: "POST",
        body: JSON.stringify({
          instructor_id: data.get("instructor_id"),
          effective_from: data.get("effective_from") ? toISO(data.get("effective_from")) : null,
          comment: emptyToNull(data.get("comment")),
        }),
      });
      await loadResources();
      render();
      toast("success", "Авто назначено.");
    });
  }

  async function saveDrafts() {
    if (state.view === "super-admin") {
      await withBusy(async function () {
        await Promise.all(state.schools.filter(function (item) { return isDirty("schools", item.id); }).map(function (item) {
          const draft = state.drafts.schools[item.id];
          return saas("/schools/" + item.id, {
            method: "PATCH",
            body: JSON.stringify({ name: draft.name, status: draft.status }),
          });
        }));

        await Promise.all(state.branches.filter(function (item) { return isDirty("branches", item.id); }).map(function (item) {
          const draft = state.drafts.branches[item.id];
          return saas("/branches/" + item.id, {
            method: "PATCH",
            body: JSON.stringify({ school_id: draft.school_id, name: draft.name, address: draft.address, status: draft.status }),
          });
        }));

        await Promise.all(state.users.filter(function (item) { return isDirty("users", item.id); }).map(function (item) {
          const draft = state.drafts.users[item.id];
          return iam("/admin/users/" + item.id, {
            method: "PATCH",
            body: JSON.stringify({ email: draft.email, role: draft.role, status: draft.status, is_active: draft.is_active }),
          });
        }));

        await loadSaas();
        render();
        toast("success", "Изменения сохранены.");
      });
      return;
    }

    await withBusy(async function () {
      await Promise.all(state.cars.filter(function (item) { return isDirty("cars", item.id); }).map(function (item) {
        const draft = state.drafts.cars[item.id];
        return resources("/resources/cars/" + item.id, {
          method: "PATCH",
          body: JSON.stringify({
            branch_id: draft.branch_id,
            instructor_id: draft.instructor_id,
            plate_number: draft.plate_number,
            brand: draft.brand,
            model: draft.model,
            color: draft.color,
            status: draft.status,
          }),
        });
      }));

      await Promise.all(state.instructors.filter(function (item) { return isDirty("instructors", item.id); }).map(function (item) {
        const draft = state.drafts.instructors[item.id];
        return resources("/resources/instructors/" + item.id, {
          method: "PATCH",
          body: JSON.stringify({
            branch_id: draft.branch_id,
            full_name: draft.full_name,
            phone: draft.phone,
            email: draft.email,
            status: draft.status,
          }),
        });
      }));

      if (availabilityDirty() && state.selectedResourceId) {
        const path = state.selectedResourceType === "cars"
          ? "/resources/cars/" + state.selectedResourceId + "/availability"
          : "/resources/instructors/" + state.selectedResourceId + "/availability";

        await resources(path, {
          method: "PUT",
          body: JSON.stringify({
            windows: state.drafts.availability.map(function (window) {
              return {
                type: window.type,
                start_at: toISO(window.start_at),
                end_at: toISO(window.end_at),
                comment: emptyToNull(window.comment),
              };
            }),
          }),
        });
      }

      await loadResources();
      render();
      toast("success", "Изменения сохранены.");
    });
  }

  async function discardDrafts() {
    await withBusy(async function () {
      if (state.view === "super-admin") {
        await loadSaas();
      } else {
        await loadResources();
      }
      render();
      toast("success", "Черновики сброшены.");
    });
  }

  async function withBusy(fn) {
    state.busy = true;
    state.toast = null;
    render();
    try {
      await fn();
    } catch (error) {
      showError(error);
    } finally {
      state.busy = false;
      render();
    }
  }

  function schoolDraft(item) {
    return { name: item.name || "", status: item.status || "active" };
  }

  function branchDraft(item) {
    return { school_id: item.school_id, name: item.name || "", address: item.address, status: item.status || "active" };
  }

  function userDraft(item) {
    return { email: item.email || "", role: item.role || "admin", status: item.status || "active", is_active: !!item.is_active };
  }

  function carDraft(item) {
    return { branch_id: item.branch_id, instructor_id: item.instructor_id || null, plate_number: item.plate_number || "", brand: item.brand, model: item.model, color: item.color, status: item.status || "active" };
  }

  function instructorDraft(item) {
    return { branch_id: item.branch_id, full_name: item.full_name || "", email: item.email, phone: item.phone, status: item.status || "active" };
  }

  function availabilityDraft(item) {
    return { id: item.id, type: item.type, start_at: toLocal(item.start_at), end_at: toLocal(item.end_at), comment: item.comment || "" };
  }

  function buildDraftMap(items, mapper) {
    const result = {};
    items.forEach(function (item) { result[item.id] = mapper(item); });
    return result;
  }

  function currentResourceCollection() {
    return state.selectedResourceType === "cars" ? state.cars : state.instructors;
  }

  function branchSelect(name, value, includePlaceholder, scope, id) {
    const attrs = [];
    if (name) attrs.push('name="' + esc(name) + '"');
    if (scope) attrs.push('data-scope="' + esc(scope) + '"');
    if (id) attrs.push('data-id="' + esc(id) + '"');
    if (scope) attrs.push('data-field="branch_id"');

    let html = "<select " + attrs.join(" ") + ">";
    if (includePlaceholder) {
      html += '<option value="">Выберите филиал</option>';
    }
    html += state.branches.map(function (branch) {
      return '<option value="' + esc(branch.id) + '"' + (String(branch.id) === String(value) ? " selected" : "") + ">" + esc(branch.name) + "</option>";
    }).join("");
    html += "</select>";
    return html;
  }

  function options(values, selected) {
    return values.map(function (value) {
      return '<option value="' + esc(value) + '"' + (String(value) === String(selected) ? " selected" : "") + ">" + esc(value) + "</option>";
    }).join("");
  }

  function saveSummary() {
    const counters = {
      schools: state.schools.filter(function (item) { return isDirty("schools", item.id); }).length,
      branches: state.branches.filter(function (item) { return isDirty("branches", item.id); }).length,
      users: state.users.filter(function (item) { return isDirty("users", item.id); }).length,
      cars: state.cars.filter(function (item) { return isDirty("cars", item.id); }).length,
      instructors: state.instructors.filter(function (item) { return isDirty("instructors", item.id); }).length,
      availability: availabilityDirty() ? 1 : 0,
    };

    const count = counters.schools + counters.branches + counters.users + counters.cars + counters.instructors + counters.availability;
    const parts = [];
    if (counters.schools) parts.push("школы: " + counters.schools);
    if (counters.branches) parts.push("филиалы: " + counters.branches);
    if (counters.users) parts.push("пользователи: " + counters.users);
    if (counters.cars) parts.push("авто: " + counters.cars);
    if (counters.instructors) parts.push("инструкторы: " + counters.instructors);
    if (counters.availability) parts.push("availability");

    return {
      count: count,
      label: parts.length ? parts.join(" · ") : "Все изменения синхронизированы",
    };
  }

  function refreshSaveDock() {
    const dock = document.getElementById("save-dock");
    if (!dock) return;
    const summary = saveSummary();
    dock.classList.toggle("visible", summary.count > 0);
  }

  function availabilityDirty() {
    return JSON.stringify(state.availability.map(availabilityDraft)) !== JSON.stringify(state.drafts.availability);
  }

  function isDirty(scope, id) {
    const draft = state.drafts[scope][id];
    if (!draft) return false;

    let original = null;
    if (scope === "schools") original = schoolDraft(findById(state.schools, id));
    if (scope === "branches") original = branchDraft(findById(state.branches, id));
    if (scope === "users") original = userDraft(findById(state.users, id));
    if (scope === "cars") original = carDraft(findById(state.cars, id));
    if (scope === "instructors") original = instructorDraft(findById(state.instructors, id));
    return JSON.stringify(draft) !== JSON.stringify(original);
  }

  function branchName(id) {
    const branch = findById(state.branches, id);
    return branch ? branch.name : id;
  }

  function findById(items, id) {
    return items.find(function (item) { return item.id === id; }) || null;
  }

  function readableRole(role) {
    if (role === "super_admin") return "Super Admin";
    if (role === "owner") return "Owner";
    if (role === "admin") return "Admin";
    return "Instructor";
  }

  function formatDateTime(value) {
    if (!value) return "—";
    return new Date(value).toLocaleString("ru-RU", {
      day: "2-digit",
      month: "2-digit",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  }

  function shortWindow(startAt, endAt) {
    return startAt.slice(0, 10) + " " + startAt.slice(11, 16) + " - " + endAt.slice(11, 16);
  }

  function toLocal(value) {
    if (!value) return "";
    const date = new Date(value);
    const offset = date.getTimezoneOffset() * 60000;
    return new Date(date.getTime() - offset).toISOString().slice(0, 16);
  }

  function toISO(value) {
    return value ? new Date(value).toISOString() : null;
  }

  function nextLocal(hour) {
    const date = new Date();
    date.setHours(hour, 0, 0, 0);
    return toLocal(date.toISOString());
  }

  function emptyToNull(value) {
    return value === "" || value === null || value === undefined ? null : value;
  }

  function nullableField(field) {
    return ["address", "brand", "model", "color", "email", "phone", "comment", "instructor_id"].indexOf(field) >= 0;
  }

  function toast(type, message) {
    state.toast = { type: type, message: message };
  }

  function showError(error) {
    toast("error", error && error.message ? error.message : "Не удалось выполнить действие.");
    render();
  }

  function esc(value) {
    return String(value === null || value === undefined ? "" : value)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&#039;");
  }

  function escAttr(value) {
    return esc(value);
  }
})();
