import { useEffect, useMemo, useState } from "react";

const STORAGE_KEY = "magistral.session";

const schoolStatusOptions = ["active", "inactive", "suspended"];
const branchStatusOptions = ["active", "inactive"];
const userRoleOptions = ["admin", "owner", "instructor"];
const userStatusOptions = ["active", "inactive", "suspended"];
const carStatusOptions = ["active", "maintenance", "inactive"];
const instructorStatusOptions = ["active", "inactive", "on_leave"];
const availabilityTypes = ["working_hours", "vacation", "blackout"];
const recurringWeekdayOptions = [
  ["monday", "Пн", 1],
  ["tuesday", "Вт", 2],
  ["wednesday", "Ср", 3],
  ["thursday", "Чт", 4],
  ["friday", "Пт", 5],
  ["saturday", "Сб", 6],
  ["sunday", "Вс", 0],
];
const studentStatusOptions = ["active", "inactive", "graduated", "paused"];
const transmissionTypeOptions = ["manual", "automatic"];
const studentPreferredWeekdayOptions = ["monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"];
const studentPreferredTimeWindowOptions = ["morning", "daytime", "evening"];
const lessonStatusOptions = ["scheduled", "completed", "cancelled", "missed"];
const paymentTypeOptions = ["cash", "card", "transfer", "other"];
const planningReplaceStrategies = ["append", "merge", "replace_generated_only"];

function readSession() {
  try {
    return JSON.parse(localStorage.getItem(STORAGE_KEY) || "null");
  } catch {
    return null;
  }
}

function writeSession(session) {
  if (session) {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(session));
  } else {
    localStorage.removeItem(STORAGE_KEY);
  }
}

function toSession(data) {
  return {
    accessToken: data.access_token,
    refreshToken: data.refresh_token,
    expiresAt: data.expires_at,
    identity: data.identity,
  };
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

function formatMoney(value) {
  const amount = Number(value || 0);
  return new Intl.NumberFormat("ru-RU", {
    style: "currency",
    currency: "RUB",
    minimumFractionDigits: 2,
  }).format(amount);
}

function asLocalDateTime(value) {
  if (!value) return "";
  const date = new Date(value);
  const shifted = new Date(date.getTime() - date.getTimezoneOffset() * 60000);
  return shifted.toISOString().slice(0, 16);
}

function toIso(value) {
  if (!value) return null;
  return new Date(value).toISOString();
}

function todayDateInput() {
  const date = new Date();
  const shifted = new Date(date.getTime() - date.getTimezoneOffset() * 60000);
  return shifted.toISOString().slice(0, 10);
}

function toLocalDateTimeInput(date) {
  const shifted = new Date(date.getTime() - date.getTimezoneOffset() * 60000);
  return shifted.toISOString().slice(0, 16);
}

function normalizeNullable(value) {
  return value === "" || value === undefined ? null : value;
}

function parseCsv(value) {
  return String(value || "")
    .split(/[,\n]/)
    .map((item) => item.trim())
    .filter(Boolean);
}

function parseMultiValue(form, name) {
  return form
    .getAll(name)
    .map((item) => String(item || "").trim())
    .filter(Boolean);
}

function parsePlaceholders(text) {
  return String(text || "")
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean)
    .reduce((acc, line) => {
      const [key, ...rest] = line.split("=");
      if (key && rest.length) {
        acc[key.trim()] = rest.join("=").trim();
      }
      return acc;
    }, {});
}

function fileToBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => {
      const result = String(reader.result || "");
      const base64 = result.includes(",") ? result.split(",").pop() : result;
      resolve(base64 || "");
    };
    reader.onerror = () => reject(reader.error || new Error("Не удалось прочитать файл"));
    reader.readAsDataURL(file);
  });
}

function downloadBase64File(fileName, mimeType, contentBase64) {
  const binary = atob(contentBase64);
  const bytes = Uint8Array.from(binary, (char) => char.charCodeAt(0));
  const blob = new Blob([bytes], { type: mimeType || "application/octet-stream" });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = fileName || "document.bin";
  link.click();
  URL.revokeObjectURL(url);
}

function nextDateTime(hoursOffset = 1) {
  const date = new Date();
  date.setHours(date.getHours() + hoursOffset, 0, 0, 0);
  return asLocalDateTime(date.toISOString());
}

function roleName(role) {
  switch (role) {
    case "super_admin":
      return "Super Admin";
    case "admin":
      return "Admin";
    case "owner":
      return "Owner";
    case "instructor":
      return "Instructor";
    default:
      return role || "Unknown";
  }
}

function useApi(session, setSession, setToast) {
  return async function request(base, path, options = {}, retry = true) {
    const headers = { ...(options.headers || {}) };
    if (!(options.body instanceof FormData)) {
      headers["Content-Type"] = headers["Content-Type"] || "application/json";
    }
    if (session?.accessToken) {
      headers.Authorization = `Bearer ${session.accessToken}`;
    }

    const response = await fetch(base + path, { ...options, headers });

    if (response.status === 401 && retry && session?.refreshToken) {
      const refreshResponse = await fetch("/api/iam/auth/refresh", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ refresh_token: session.refreshToken }),
      });

      if (!refreshResponse.ok) {
        setSession(null);
        writeSession(null);
        throw new Error("Сессия истекла");
      }

      const refreshPayload = await refreshResponse.json();
      const nextSession = toSession(refreshPayload.data);
      setSession(nextSession);
      writeSession(nextSession);
      return request(base, path, options, false);
    }

    const payload = response.status === 204 ? null : await response.json().catch(() => null);
    if (!response.ok) {
      const message = payload?.error?.message || payload?.message || `HTTP ${response.status}`;
      setToast({ type: "error", message });
      throw new Error(message);
    }

    return payload;
  };
}

function App() {
  const [session, setSession] = useState(() => readSession());
  const [me, setMe] = useState(null);
  const [view, setView] = useState("login");
  const [workspaceTab, setWorkspaceTab] = useState("overview");
  const [busy, setBusy] = useState(false);
  const [toast, setToast] = useState(null);

  const [schools, setSchools] = useState([]);
  const [selectedSchoolId, setSelectedSchoolId] = useState("");
  const [branches, setBranches] = useState([]);
  const [users, setUsers] = useState([]);

  const [cars, setCars] = useState([]);
  const [instructors, setInstructors] = useState([]);
  const [selectedResourceType, setSelectedResourceType] = useState("cars");
  const [selectedResourceId, setSelectedResourceId] = useState("");
  const [availability, setAvailability] = useState([]);
  const [availabilityDrafts, setAvailabilityDrafts] = useState([]);

  const [students, setStudents] = useState([]);
  const [studyPrograms, setStudyPrograms] = useState([]);
  const [groups, setGroups] = useState([]);
  const [lessons, setLessons] = useState([]);
  const [payments, setPayments] = useState([]);
  const [documents, setDocuments] = useState([]);
  const [documentTemplates, setDocumentTemplates] = useState([]);
  const [reportingDashboard, setReportingDashboard] = useState(null);
  const [reportingStudents, setReportingStudents] = useState(null);
  const [reportingLessons, setReportingLessons] = useState(null);
  const [reportingFinance, setReportingFinance] = useState(null);
  const [reportingInstructorWorkload, setReportingInstructorWorkload] = useState([]);
  const [reportingCarsUtilization, setReportingCarsUtilization] = useState([]);
  const [reportingFilters, setReportingFilters] = useState({ branch_id: "", group_id: "", month: "" });
  const [selectedStudentId, setSelectedStudentId] = useState("");
  const [studentBalance, setStudentBalance] = useState(null);
  const [studentGroupFilter, setStudentGroupFilter] = useState("");
  const [editingStudentId, setEditingStudentId] = useState("");
  const [editingProgramId, setEditingProgramId] = useState("");
  const [editingGroupId, setEditingGroupId] = useState("");
  const [editingLessonId, setEditingLessonId] = useState("");
  const [editingPaymentId, setEditingPaymentId] = useState("");
  const [editingTemplateId, setEditingTemplateId] = useState("");
  const [documentDetails, setDocumentDetails] = useState(null);

  const request = useApi(session, setSession, setToast);
  const iam = (path, options) => request("/api/iam", path, options);
  const saas = (path, options) => request("/api/saas", path, options);
  const resources = (path, options) => request("/api/resources", path, options);
  const studentsApi = (path, options) => request("/api/students", path, options);
  const lessonsApi = (path, options) => request("/api/lessons", path, options);
  const paymentsApi = (path, options) => request("/api/payments", path, options);
  const documentsApi = (path, options) => request("/api/documents", path, options);
  const reportingApi = (path, options) => request("/api/reporting", path, options);

  const isSuperAdmin = me?.role === "super_admin";
  const canWriteWorkspace = me?.role === "admin";
  const canSeeWorkspace = Boolean(me && me.role !== "super_admin");

  const selectedSchool = useMemo(
    () => schools.find((item) => item.id === selectedSchoolId) || null,
    [schools, selectedSchoolId]
  );
  const editingStudent = useMemo(
    () => students.find((item) => item.id === editingStudentId) || null,
    [students, editingStudentId]
  );
  const editingProgram = useMemo(
    () => studyPrograms.find((item) => item.id === editingProgramId) || null,
    [studyPrograms, editingProgramId]
  );
  const editingGroup = useMemo(
    () => groups.find((item) => item.id === editingGroupId) || null,
    [groups, editingGroupId]
  );
  const editingLesson = useMemo(
    () => lessons.find((item) => item.id === editingLessonId) || null,
    [lessons, editingLessonId]
  );
  const editingPayment = useMemo(
    () => payments.find((item) => item.id === editingPaymentId) || null,
    [payments, editingPaymentId]
  );
  const editingTemplate = useMemo(
    () => documentTemplates.find((item) => item.id === editingTemplateId) || null,
    [documentTemplates, editingTemplateId]
  );
  const filteredStudents = useMemo(
    () => (studentGroupFilter ? students.filter((item) => item.group_id === studentGroupFilter) : students),
    [studentGroupFilter, students]
  );

  const selectedResourceCollection = selectedResourceType === "cars" ? cars : instructors;
  const selectedResource = selectedResourceCollection.find((item) => item.id === selectedResourceId) || null;

  const branchNameById = useMemo(
    () => Object.fromEntries(branches.map((item) => [item.id, item.name])),
    [branches]
  );
  const studentNameById = useMemo(
    () => Object.fromEntries(students.map((item) => [item.id, item.full_name])),
    [students]
  );
  const studyProgramNameById = useMemo(
    () => Object.fromEntries(studyPrograms.map((item) => [item.id, item.name])),
    [studyPrograms]
  );
  const groupNameById = useMemo(
    () => Object.fromEntries(groups.map((item) => [item.id, item.name])),
    [groups]
  );
  const instructorNameById = useMemo(
    () => Object.fromEntries(instructors.map((item) => [item.id, item.full_name])),
    [instructors]
  );
  const carNameById = useMemo(
    () =>
      Object.fromEntries(
        cars.map((item) => [
          item.id,
          [item.plate_number, item.brand, item.model].filter(Boolean).join(" • "),
        ])
      ),
    [cars]
  );

  useEffect(() => {
    if (!session) {
      setMe(null);
      setView("login");
      return;
    }

    let cancelled = false;
    async function bootstrap() {
      setBusy(true);
      try {
        const payload = await iam("/auth/me");
        if (cancelled) return;
        const nextMe = payload.data;
        setMe(nextMe);
        if (nextMe.role === "super_admin") {
          setView("super-admin");
          await loadSuperAdminData();
        } else {
          setView("workspace");
          await loadWorkspaceData(nextMe);
        }
      } catch (_error) {
        if (!cancelled) {
          setSession(null);
          writeSession(null);
        }
      } finally {
        if (!cancelled) setBusy(false);
      }
    }

    void bootstrap();
    return () => {
      cancelled = true;
    };
  }, [session]);

  useEffect(() => {
    if (!canSeeWorkspace || !selectedStudentId) {
      setStudentBalance(null);
      return;
    }
    void loadStudentBalance(selectedStudentId);
  }, [selectedStudentId, canSeeWorkspace]);

  async function withBusy(task) {
    setBusy(true);
    setToast(null);
    try {
      await task();
    } catch (_error) {
      return;
    } finally {
      setBusy(false);
    }
  }

  async function loadSuperAdminData(nextSchoolId) {
    const schoolPayload = await saas("/schools?limit=100&offset=0");
    const nextSchools = schoolPayload.data || [];
    setSchools(nextSchools);

    const schoolId =
      nextSchoolId && nextSchools.some((item) => item.id === nextSchoolId)
        ? nextSchoolId
        : nextSchools[0]?.id || "";
    setSelectedSchoolId(schoolId);

    if (!schoolId) {
      setBranches([]);
      setUsers([]);
      return;
    }

    const [branchPayload, userPayload] = await Promise.all([
      saas(`/branches?school_id=${encodeURIComponent(schoolId)}&limit=100&offset=0`),
      iam(`/admin/users?school_id=${encodeURIComponent(schoolId)}&limit=100&offset=0`),
    ]);

    setBranches(branchPayload.data || []);
    setUsers(userPayload.data?.users || []);
  }

  async function loadWorkspaceData(currentMe = me) {
    const schoolId = currentMe?.school_id;
    if (!schoolId) return;

    const [
      branchPayload,
      carsPayload,
      instructorsPayload,
      studyProgramsPayload,
      studentsPayload,
      groupsPayload,
      lessonsPayload,
      paymentsPayload,
      documentsPayload,
      templatesPayload,
    ] = await Promise.all([
      saas(`/branches?school_id=${encodeURIComponent(schoolId)}&limit=100&offset=0`),
      resources("/resources/cars?limit=100&offset=0"),
      resources("/resources/instructors?limit=100&offset=0"),
      studentsApi("/study-programs?limit=100&offset=0"),
      studentsApi("/students?limit=100&offset=0"),
      studentsApi("/groups?limit=100&offset=0"),
      lessonsApi("/lessons?limit=100&offset=0"),
      paymentsApi("/payments?limit=100&offset=0"),
      documentsApi("/documents?limit=100&offset=0"),
      documentsApi("/document-templates?limit=100&offset=0"),
    ]);

    const nextBranches = branchPayload.data || [];
    const nextCars = carsPayload.data || [];
    const nextInstructors = instructorsPayload.data || [];
    const nextStudyPrograms = studyProgramsPayload.data || [];
    const nextStudents = studentsPayload.data || [];
    const nextGroups = groupsPayload.data || [];
    const nextLessons = lessonsPayload.data || [];
    const nextPayments = paymentsPayload.data || [];
    const nextDocuments = documentsPayload.data || [];
    const nextTemplates = templatesPayload.data || [];

    setBranches(nextBranches);
    setCars(nextCars);
    setInstructors(nextInstructors);
    setStudyPrograms(nextStudyPrograms);
    setStudents(nextStudents);
    setGroups(nextGroups);
    setLessons(nextLessons);
    setPayments(nextPayments);
    setDocuments(nextDocuments);
    setDocumentTemplates(nextTemplates);

    const currentCollection = selectedResourceType === "cars" ? nextCars : nextInstructors;
    const resourceId =
      selectedResourceId && currentCollection.some((item) => item.id === selectedResourceId)
        ? selectedResourceId
        : nextCars[0]?.id || nextInstructors[0]?.id || "";
    const resourceType =
      selectedResourceType === "cars"
        ? nextCars.length
          ? "cars"
          : "instructors"
        : nextInstructors.length
          ? "instructors"
          : "cars";

    setSelectedResourceType(resourceType);
    setSelectedResourceId(resourceId);

    const nextStudentId =
      selectedStudentId && nextStudents.some((item) => item.id === selectedStudentId)
        ? selectedStudentId
        : nextStudents[0]?.id || "";
    setSelectedStudentId(nextStudentId);

    if (resourceId) {
      await loadAvailability(resourceType, resourceId);
    } else {
      setAvailability([]);
      setAvailabilityDrafts([]);
    }

    await loadReportingData();
  }

  async function loadReportingData() {
    const query = buildReportingQuery(reportingFilters);
    const results = await Promise.allSettled([
      reportingApi(`/reports/dashboard${query}`),
      reportingApi(`/reports/students${query}`),
      reportingApi(`/reports/lessons${query}`),
      reportingApi(`/reports/finance${query}`),
      reportingApi(`/reports/instructors/workload${query}`),
      reportingApi(`/reports/cars/utilization${query}`),
    ]);

    setReportingDashboard(results[0].status === "fulfilled" ? results[0].value?.data || null : null);
    setReportingStudents(results[1].status === "fulfilled" ? results[1].value?.data || null : null);
    setReportingLessons(results[2].status === "fulfilled" ? results[2].value?.data || null : null);
    setReportingFinance(results[3].status === "fulfilled" ? results[3].value?.data || null : null);
    setReportingInstructorWorkload(results[4].status === "fulfilled" ? results[4].value?.data || [] : []);
    setReportingCarsUtilization(results[5].status === "fulfilled" ? results[5].value?.data || [] : []);
  }

  function buildReportingQuery(filters) {
    const params = new URLSearchParams();
    if (filters.branch_id) params.set("branch_id", filters.branch_id);
    if (filters.group_id) params.set("group_id", filters.group_id);
    if (filters.month) params.set("month", filters.month);
    const value = params.toString();
    return value ? `?${value}` : "";
  }

  async function applyReportingFilters(event) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const nextFilters = {
      branch_id: String(form.get("branch_id") || ""),
      group_id: String(form.get("group_id") || ""),
      month: String(form.get("month") || ""),
    };
    setReportingFilters(nextFilters);
    await withBusy(async () => {
      const query = buildReportingQuery(nextFilters);
      const results = await Promise.allSettled([
        reportingApi(`/reports/dashboard${query}`),
        reportingApi(`/reports/students${query}`),
        reportingApi(`/reports/lessons${query}`),
        reportingApi(`/reports/finance${query}`),
        reportingApi(`/reports/instructors/workload${query}`),
        reportingApi(`/reports/cars/utilization${query}`),
      ]);

      setReportingDashboard(results[0].status === "fulfilled" ? results[0].value?.data || null : null);
      setReportingStudents(results[1].status === "fulfilled" ? results[1].value?.data || null : null);
      setReportingLessons(results[2].status === "fulfilled" ? results[2].value?.data || null : null);
      setReportingFinance(results[3].status === "fulfilled" ? results[3].value?.data || null : null);
      setReportingInstructorWorkload(results[4].status === "fulfilled" ? results[4].value?.data || [] : []);
      setReportingCarsUtilization(results[5].status === "fulfilled" ? results[5].value?.data || [] : []);
    });
  }

  async function loadAvailability(resourceType, resourceId) {
    if (!resourceId) {
      setAvailability([]);
      setAvailabilityDrafts([]);
      return;
    }
    const payload = await resources(
      resourceType === "cars"
        ? `/resources/cars/${resourceId}/availability`
        : `/resources/instructors/${resourceId}/availability`
    );
    const nextAvailability = payload.data || [];
    setAvailability(nextAvailability);
    setAvailabilityDrafts(
      nextAvailability.map((item) => ({
        id: item.id,
        type: item.type,
        start_at: asLocalDateTime(item.start_at),
        end_at: asLocalDateTime(item.end_at),
        comment: item.comment || "",
      }))
    );
  }

  async function loadStudentBalance(studentId) {
    try {
      const payload = await paymentsApi(`/payments/students/${studentId}/balance`);
      setStudentBalance(payload.data || null);
    } catch (_error) {
      setStudentBalance(null);
    }
  }

  async function handleLogin(event) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    await withBusy(async () => {
      const payload = await iam("/auth/login", {
        method: "POST",
        body: JSON.stringify({
          email: form.get("email"),
          password: form.get("password"),
        }),
      });
      const nextSession = toSession(payload.data);
      setSession(nextSession);
      writeSession(nextSession);
      setToast({ type: "success", message: "Сессия открыта." });
    });
  }

  function handleLogout() {
    setSession(null);
    writeSession(null);
    setMe(null);
    setView("login");
    setWorkspaceTab("overview");
    setToast(null);
  }

  async function handleDelete(requestFn, successMessage) {
    await withBusy(async () => {
      await requestFn();
      setToast({ type: "success", message: successMessage });
      if (isSuperAdmin) {
        await loadSuperAdminData(selectedSchoolId);
      } else {
        await loadWorkspaceData();
      }
    });
  }

  async function createSchool(event) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    await withBusy(async () => {
      await saas("/schools", {
        method: "POST",
        body: JSON.stringify({
          name: form.get("name"),
          status: form.get("status"),
        }),
      });
      event.currentTarget.reset();
      setToast({ type: "success", message: "Школа создана." });
      await loadSuperAdminData(selectedSchoolId);
    });
  }

  async function bootstrapSchool(event) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    await withBusy(async () => {
      await saas("/admin/bootstrap-school", {
        method: "POST",
        body: JSON.stringify({
          name: form.get("name"),
          status: form.get("status"),
          admin_email: form.get("admin_email"),
          password: form.get("password"),
        }),
      });
      event.currentTarget.reset();
      setToast({ type: "success", message: "Школа и первый admin созданы." });
      await loadSuperAdminData(selectedSchoolId);
    });
  }

  async function createBranch(event) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    await withBusy(async () => {
      await saas("/branches", {
        method: "POST",
        body: JSON.stringify({
          school_id: selectedSchoolId,
          name: form.get("name"),
          address: normalizeNullable(form.get("address")),
          status: form.get("status"),
        }),
      });
      event.currentTarget.reset();
      setToast({ type: "success", message: "Филиал создан." });
      await loadSuperAdminData(selectedSchoolId);
    });
  }

  async function createUser(event) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    await withBusy(async () => {
      await iam("/admin/users", {
        method: "POST",
        body: JSON.stringify({
          school_id: selectedSchoolId,
          email: form.get("email"),
          password: form.get("password"),
          role: form.get("role"),
          status: form.get("status"),
        }),
      });
      event.currentTarget.reset();
      setToast({ type: "success", message: "Пользователь создан." });
      await loadSuperAdminData(selectedSchoolId);
    });
  }

  async function createCar(event) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    await withBusy(async () => {
      await resources("/resources/cars", {
        method: "POST",
        body: JSON.stringify({
          branch_id: form.get("branch_id"),
          instructor_id: normalizeNullable(form.get("instructor_id")),
          plate_number: form.get("plate_number"),
          brand: normalizeNullable(form.get("brand")),
          model: normalizeNullable(form.get("model")),
          color: normalizeNullable(form.get("color")),
          transmission: form.get("transmission"),
          status: form.get("status"),
        }),
      });
      event.currentTarget.reset();
      setToast({ type: "success", message: "Автомобиль создан." });
      await loadWorkspaceData();
    });
  }

  async function createInstructor(event) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    await withBusy(async () => {
      await resources("/resources/instructors", {
        method: "POST",
        body: JSON.stringify({
          branch_id: form.get("branch_id"),
          full_name: form.get("full_name"),
          email: normalizeNullable(form.get("email")),
          phone: normalizeNullable(form.get("phone")),
          status: form.get("status"),
        }),
      });
      event.currentTarget.reset();
      setToast({ type: "success", message: "Инструктор создан." });
      await loadWorkspaceData();
    });
  }

  async function updateSelectedResource(event) {
    event.preventDefault();
    if (!selectedResourceId) return;
    const form = new FormData(event.currentTarget);
    const path =
      selectedResourceType === "cars"
        ? `/resources/cars/${selectedResourceId}`
        : `/resources/instructors/${selectedResourceId}`;
    const body =
      selectedResourceType === "cars"
        ? {
            branch_id: form.get("branch_id"),
            instructor_id: normalizeNullable(form.get("instructor_id")),
            plate_number: form.get("plate_number"),
            brand: normalizeNullable(form.get("brand")),
            model: normalizeNullable(form.get("model")),
            color: normalizeNullable(form.get("color")),
            transmission: form.get("transmission"),
            status: form.get("status"),
          }
        : {
            branch_id: form.get("branch_id"),
            full_name: form.get("full_name"),
            email: normalizeNullable(form.get("email")),
            phone: normalizeNullable(form.get("phone")),
            status: form.get("status"),
          };

    await withBusy(async () => {
      await resources(path, {
        method: "PATCH",
        body: JSON.stringify(body),
      });
      setToast({ type: "success", message: "Resource updated." });
      await loadWorkspaceData();
    });
  }

  function addRecurringAvailability(event) {
    event.preventDefault();
    if (!selectedResourceId) return;

    const form = new FormData(event.currentTarget);
    const startDateValue = String(form.get("start_date") || "");
    const startTimeValue = String(form.get("start_time") || "");
    const endTimeValue = String(form.get("end_time") || "");
    const repeatMonths = Math.max(1, Math.min(12, Number(form.get("repeat_months") || 3)));
    const selectedWeekdays = new Set(
      recurringWeekdayOptions
        .filter(([value]) => form.get(`weekday_${value}`) === "on")
        .map(([, , dayIndex]) => dayIndex)
    );

    if (!startDateValue || !startTimeValue || !endTimeValue || !selectedWeekdays.size) {
      setToast({ type: "error", message: "Заполните дату, время и дни недели для повтора." });
      return;
    }

    const startDate = new Date(`${startDateValue}T00:00:00`);
    const repeatUntil = new Date(startDate);
    repeatUntil.setMonth(repeatUntil.getMonth() + repeatMonths);

    const nextWindows = [];
    const cursor = new Date(startDate);
    while (cursor < repeatUntil) {
      if (selectedWeekdays.has(cursor.getDay())) {
        const datePart = toLocalDateTimeInput(cursor).slice(0, 10);
        const startAt = new Date(`${datePart}T${startTimeValue}:00`);
        const endAt = new Date(`${datePart}T${endTimeValue}:00`);
        if (endAt <= startAt) {
          setToast({ type: "error", message: "Время окончания должно быть позже времени начала." });
          return;
        }
        nextWindows.push({
          id: `recurring-${Date.now()}-${nextWindows.length}`,
          type: String(form.get("type") || "working_hours"),
          start_at: toLocalDateTimeInput(startAt),
          end_at: toLocalDateTimeInput(endAt),
          comment: String(form.get("comment") || ""),
        });
      }
      cursor.setDate(cursor.getDate() + 1);
    }

    setAvailabilityDrafts((current) => [...current, ...nextWindows]);
    event.currentTarget.reset();
    setToast({ type: "success", message: `Добавлено повторяющихся окон: ${nextWindows.length}.` });
  }

  async function saveAvailability() {
    await withBusy(async () => {
      await resources(
        selectedResourceType === "cars"
          ? `/resources/cars/${selectedResourceId}/availability`
          : `/resources/instructors/${selectedResourceId}/availability`,
        {
          method: "PUT",
          body: JSON.stringify({
            windows: availabilityDrafts.map((item) => ({
              type: item.type,
              start_at: toIso(item.start_at),
              end_at: toIso(item.end_at),
              comment: normalizeNullable(item.comment),
            })),
          }),
        }
      );
      setToast({ type: "success", message: "Availability обновлён." });
      await loadAvailability(selectedResourceType, selectedResourceId);
    });
  }

  async function createStudent(event) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    await withBusy(async () => {
      await studentsApi("/students", {
        method: "POST",
        body: JSON.stringify({
          branch_id: form.get("branch_id"),
          group_id: normalizeNullable(form.get("group_id")),
          program_id: form.get("program_id"),
          full_name: form.get("full_name"),
          phone: form.get("phone"),
          status: form.get("status"),
          preferred_weekdays: parseMultiValue(form, "preferred_weekdays"),
          preferred_time_windows: parseMultiValue(form, "preferred_time_windows"),
          manual_hours_completed: Number(form.get("manual_hours_completed") || 0),
        }),
      });
      event.currentTarget.reset();
      setToast({ type: "success", message: "Ученик создан." });
      await loadWorkspaceData();
    });
  }

  async function createStudyProgram(event) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    await withBusy(async () => {
      await studentsApi("/study-programs", {
        method: "POST",
        body: JSON.stringify({
          name: form.get("name"),
          transmission: form.get("transmission"),
          total_hours: Number(form.get("total_hours") || 0),
        }),
      });
      event.currentTarget.reset();
      setToast({ type: "success", message: "Study program created." });
      await loadWorkspaceData();
    });
  }

  async function createGroup(event) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    await withBusy(async () => {
      await studentsApi("/groups", {
        method: "POST",
        body: JSON.stringify({
          branch_id: form.get("branch_id"),
          name: form.get("name"),
        }),
      });
      event.currentTarget.reset();
      setToast({ type: "success", message: "Группа создана." });
      await loadWorkspaceData();
    });
  }

  async function updateStudent(event) {
    event.preventDefault();
    if (!editingStudentId) return;
    const form = new FormData(event.currentTarget);
    await withBusy(async () => {
      await studentsApi(`/students/${editingStudentId}`, {
        method: "PATCH",
        body: JSON.stringify({
          branch_id: form.get("branch_id"),
          group_id: normalizeNullable(form.get("group_id")),
          program_id: form.get("program_id"),
          full_name: form.get("full_name"),
          phone: form.get("phone"),
          status: form.get("status"),
          preferred_weekdays: parseMultiValue(form, "preferred_weekdays"),
          preferred_time_windows: parseMultiValue(form, "preferred_time_windows"),
          manual_hours_completed: Number(form.get("manual_hours_completed") || 0),
        }),
      });
      setToast({ type: "success", message: "Карточка ученика обновлена." });
      await loadWorkspaceData();
    });
  }

  async function updateStudyProgram(event) {
    event.preventDefault();
    if (!editingProgramId) return;
    const form = new FormData(event.currentTarget);
    await withBusy(async () => {
      await studentsApi(`/study-programs/${editingProgramId}`, {
        method: "PATCH",
        body: JSON.stringify({
          name: form.get("name"),
          transmission: form.get("transmission"),
          total_hours: Number(form.get("total_hours") || 0),
        }),
      });
      setToast({ type: "success", message: "Study program updated." });
      await loadWorkspaceData();
    });
  }

  async function updateGroup(event) {
    event.preventDefault();
    if (!editingGroupId) return;
    const form = new FormData(event.currentTarget);
    await withBusy(async () => {
      await studentsApi(`/groups/${editingGroupId}`, {
        method: "PATCH",
        body: JSON.stringify({
          branch_id: form.get("branch_id"),
          name: form.get("name"),
        }),
      });
      setToast({ type: "success", message: "Группа обновлена." });
      await loadWorkspaceData();
    });
  }

  async function createLesson(event) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    await withBusy(async () => {
      await lessonsApi("/lessons", {
        method: "POST",
        body: JSON.stringify({
          branch_id: normalizeNullable(form.get("branch_id")),
          student_id: form.get("student_id"),
          instructor_id: form.get("instructor_id"),
          car_id: form.get("car_id"),
          start_at: toIso(form.get("start_at")),
          end_at: toIso(form.get("end_at")),
          status: form.get("status"),
          comment: normalizeNullable(form.get("comment")),
        }),
      });
      event.currentTarget.reset();
      setToast({ type: "success", message: "Занятие создано." });
      await loadWorkspaceData();
    });
  }

  async function updateLesson(event) {
    event.preventDefault();
    if (!editingLessonId) return;
    const form = new FormData(event.currentTarget);
    await withBusy(async () => {
      await lessonsApi(`/lessons/${editingLessonId}`, {
        method: "PATCH",
        body: JSON.stringify({
          branch_id: normalizeNullable(form.get("branch_id")),
          student_id: form.get("student_id"),
          instructor_id: form.get("instructor_id"),
          car_id: form.get("car_id"),
          start_at: toIso(form.get("start_at")),
          end_at: toIso(form.get("end_at")),
          status: form.get("status"),
          comment: normalizeNullable(form.get("comment")),
        }),
      });
      setToast({ type: "success", message: "Занятие обновлено." });
      await loadWorkspaceData();
    });
  }

  async function createPayment(event) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    await withBusy(async () => {
      await paymentsApi("/payments", {
        method: "POST",
        body: JSON.stringify({
          student_id: form.get("student_id"),
          type: form.get("type"),
          amount: Number(form.get("amount")),
          comment: normalizeNullable(form.get("comment")),
        }),
      });
      event.currentTarget.reset();
      setToast({ type: "success", message: "Платёж создан." });
      await loadWorkspaceData();
    });
  }

  async function updateContract(event) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    await withBusy(async () => {
      await paymentsApi(`/payments/students/${form.get("student_id")}/contract`, {
        method: "PUT",
        body: JSON.stringify({
          contract_total: Number(form.get("contract_total")),
        }),
      });
      setToast({ type: "success", message: "Сумма договора обновлена." });
      await loadWorkspaceData();
    });
  }

  async function updatePayment(event) {
    event.preventDefault();
    if (!editingPaymentId) return;
    const form = new FormData(event.currentTarget);
    await withBusy(async () => {
      await paymentsApi(`/payments/${editingPaymentId}`, {
        method: "PATCH",
        body: JSON.stringify({
          type: form.get("type"),
          amount: Number(form.get("amount")),
          comment: normalizeNullable(form.get("comment")),
        }),
      });
      setToast({ type: "success", message: "Платёж обновлён." });
      await loadWorkspaceData();
    });
  }

  async function uploadDocument(event) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const file = form.get("file");
    if (!(file instanceof File) || !file.size) {
      setToast({ type: "error", message: "Выберите файл для загрузки." });
      return;
    }

    await withBusy(async () => {
      const contentBase64 = await fileToBase64(file);
      await documentsApi("/documents/upload", {
        method: "POST",
        body: JSON.stringify({
          student_id: form.get("student_id"),
          type: form.get("type"),
          title: form.get("title"),
          mime_type: file.type || "application/octet-stream",
          file_name: file.name,
          content_base64: contentBase64,
        }),
      });
      event.currentTarget.reset();
      setToast({ type: "success", message: "Документ загружен." });
      await loadWorkspaceData();
    });
  }

  async function generateDocument(event) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    await withBusy(async () => {
      await documentsApi("/documents/generate", {
        method: "POST",
        body: JSON.stringify({
          student_id: form.get("student_id"),
          template_id: form.get("template_id"),
          title: form.get("title"),
          file_name: normalizeNullable(form.get("file_name")),
          placeholders: parsePlaceholders(form.get("placeholders")),
        }),
      });
      event.currentTarget.reset();
      setToast({ type: "success", message: "Документ сгенерирован." });
      await loadWorkspaceData();
    });
  }

  async function createTemplate(event) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    await withBusy(async () => {
      await documentsApi("/document-templates", {
        method: "POST",
        body: JSON.stringify({
          name: form.get("name"),
          type: form.get("type"),
          content: form.get("content"),
        }),
      });
      event.currentTarget.reset();
      setToast({ type: "success", message: "Шаблон создан." });
      await loadWorkspaceData();
    });
  }

  async function updateTemplate(event) {
    event.preventDefault();
    if (!editingTemplateId) return;
    const form = new FormData(event.currentTarget);
    await withBusy(async () => {
      await documentsApi(`/document-templates/${editingTemplateId}`, {
        method: "PATCH",
        body: JSON.stringify({
          name: form.get("name"),
          type: form.get("type"),
          content: form.get("content"),
        }),
      });
      setToast({ type: "success", message: "Шаблон обновлён." });
      await loadWorkspaceData();
    });
  }

  async function requestPlanning(event) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    await withBusy(async () => {
      const payload = await lessonsApi("/planning/requests", {
        method: "POST",
        body: JSON.stringify({
          period_start: toIso(form.get("period_start")),
          period_end: toIso(form.get("period_end")),
          student_ids: parseMultiValue(form, "student_ids"),
          group_ids: parseMultiValue(form, "group_ids"),
          branch_ids: parseMultiValue(form, "branch_ids"),
          replace_strategy: form.get("replace_strategy"),
          request_id: normalizeNullable(form.get("request_id")),
        }),
      });
      event.currentTarget.reset();
      setToast({
        type: "success",
        message: `Planning request queued: ${payload.data?.request_id || "queued"}`,
      });
    });
  }

  async function downloadDocument(documentId) {
    await withBusy(async () => {
      const payload = await documentsApi(`/documents/${documentId}/download`);
      const data = payload.data;
      downloadBase64File(data.file_name, data.mime_type, data.content_base64);
      setToast({ type: "success", message: "Документ скачан." });
    });
  }

  async function inspectDocument(documentId) {
    await withBusy(async () => {
      const payload = await documentsApi(`/documents/${documentId}`);
      setDocumentDetails(payload.data || null);
      setToast({ type: "success", message: "Метаданные документа загружены." });
    });
  }

  function addAvailabilityWindow() {
    setAvailabilityDrafts((prev) => [
      ...prev,
      {
        id: `draft-${crypto.randomUUID()}`,
        type: "working_hours",
        start_at: nextDateTime(1),
        end_at: nextDateTime(3),
        comment: "",
      },
    ]);
  }

  function updateAvailabilityWindow(id, field, value) {
    setAvailabilityDrafts((prev) =>
      prev.map((item) => (item.id === id ? { ...item, [field]: value } : item))
    );
  }

  function removeAvailabilityWindow(id) {
    setAvailabilityDrafts((prev) => prev.filter((item) => item.id !== id));
  }

  function renderLogin() {
    return (
      <div className="login-layout">
        <section className="login-hero">
          <span className="hero-badge">Magistral Workspace</span>
          <h1>Единый кабинет для SaaS, ресурсов, учеников, платежей и документов.</h1>
          <p>
            Интерфейс собран под ручную проверку всех основных сервисов платформы: от
            bootstrap школы до планирования расписания через Kafka.
          </p>

          <div className="hero-strip">
            <StatCard label="Сервисы" value="9" hint="iam, saas, resources, student, lesson, payment, document, reporting, planning" />
            <StatCard label="HTTP Gateway" value="7" hint="Через nginx proxy в frontend" />
            <StatCard label="Основной режим" value="School" hint="Admin workspace для ежедневной работы" />
          </div>
        </section>

        <section className="login-panel">
          <div className="panel-head compact">
            <div>
              <span className="eyebrow">Авторизация</span>
              <h2>Вход в систему</h2>
            </div>
          </div>

          <form className="form-stack" onSubmit={handleLogin}>
            <label>
              Email
              <input name="email" type="email" placeholder="admin@example.com" required />
            </label>
            <label>
              Пароль
              <input name="password" type="password" placeholder="Passw0rd!" required />
            </label>
            <button className="primary-button big" type="submit" disabled={busy}>
              {busy ? "Входим..." : "Войти"}
            </button>
          </form>

          <div className="hint-box">
            <span className="eyebrow">Bootstrap</span>
            <strong>Super Admin по умолчанию</strong>
            <span>superadmin@example.com / Passw0rd!</span>
          </div>

          {renderToast()}
        </section>
      </div>
    );
  }

  function renderSuperAdmin() {
    return (
      <>
        <section className="hero-strip">
          <StatCard label="Школы" value={schools.length} hint="SaaS registry" />
          <StatCard label="Филиалы" value={branches.length} hint="Контекст выбранной школы" />
          <StatCard label="Пользователи" value={users.length} hint="IAM school-scoped users" />
          <StatCard label="Режим" value="SaaS" hint="Полный доступ super_admin" />
        </section>

        <section className="workspace-grid super-admin-grid">
          <div className="stack-column">
            <div className="panel">
              <div className="panel-head">
                <div>
                  <span className="eyebrow">Bootstrap</span>
                  <h3>Создать школу и первого admin</h3>
                </div>
              </div>
              <form className="form-stack" onSubmit={bootstrapSchool}>
                <input name="name" placeholder="Название школы" required />
                <select name="status" defaultValue="active">
                  {schoolStatusOptions.map((option) => (
                    <option key={option} value={option}>
                      {option}
                    </option>
                  ))}
                </select>
                <input name="admin_email" type="email" placeholder="admin@school.ru" required />
                <input name="password" type="password" placeholder="Пароль" required />
                <button className="primary-button" type="submit">
                  Bootstrap
                </button>
              </form>
            </div>

            <div className="panel">
              <div className="panel-head">
                <div>
                  <span className="eyebrow">Школы</span>
                  <h3>Создать школу вручную</h3>
                </div>
              </div>
              <form className="form-stack" onSubmit={createSchool}>
                <input name="name" placeholder="Название школы" required />
                <select name="status" defaultValue="active">
                  {schoolStatusOptions.map((option) => (
                    <option key={option} value={option}>
                      {option}
                    </option>
                  ))}
                </select>
                <button className="primary-button" type="submit">
                  Создать школу
                </button>
              </form>
            </div>
          </div>

          <div className="stack-column">
            <div className="panel">
              <div className="panel-head">
                <div>
                  <span className="eyebrow">Контекст</span>
                  <h3>Выбранная школа</h3>
                </div>
                <select
                  value={selectedSchoolId}
                  onChange={(event) => {
                    const value = event.target.value;
                    setSelectedSchoolId(value);
                    void withBusy(() => loadSuperAdminData(value));
                  }}
                >
                  {schools.map((item) => (
                    <option key={item.id} value={item.id}>
                      {item.name}
                    </option>
                  ))}
                </select>
              </div>

              {selectedSchool ? (
                <div className="detail-grid">
                  <Detail label="ID" value={selectedSchool.id} />
                  <Detail label="Статус" value={selectedSchool.status} />
                  <Detail label="Создана" value={formatDateTime(selectedSchool.created_at)} />
                  <Detail label="Обновлена" value={formatDateTime(selectedSchool.updated_at)} />
                </div>
              ) : (
                <EmptyState message="Сначала создайте школу." />
              )}
            </div>

            <div className="panel">
              <div className="panel-head">
                <div>
                  <span className="eyebrow">Филиалы</span>
                  <h3>Добавить филиал</h3>
                </div>
              </div>
              <form className="toolbar-form" onSubmit={createBranch}>
                <input name="name" placeholder="Название филиала" required disabled={!selectedSchoolId} />
                <input name="address" placeholder="Адрес" disabled={!selectedSchoolId} />
                <select name="status" defaultValue="active" disabled={!selectedSchoolId}>
                  {branchStatusOptions.map((option) => (
                    <option key={option} value={option}>
                      {option}
                    </option>
                  ))}
                </select>
                <button className="primary-button" type="submit" disabled={!selectedSchoolId}>
                  Добавить
                </button>
              </form>
            </div>

            <div className="panel">
              <div className="panel-head">
                <div>
                  <span className="eyebrow">IAM</span>
                  <h3>Создать school user</h3>
                </div>
              </div>
              <form className="toolbar-form wide" onSubmit={createUser}>
                <input name="email" type="email" placeholder="user@school.ru" required disabled={!selectedSchoolId} />
                <input name="password" type="password" placeholder="Пароль" required disabled={!selectedSchoolId} />
                <select name="role" defaultValue="admin" disabled={!selectedSchoolId}>
                  {userRoleOptions.map((option) => (
                    <option key={option} value={option}>
                      {option}
                    </option>
                  ))}
                </select>
                <select name="status" defaultValue="active" disabled={!selectedSchoolId}>
                  {userStatusOptions.map((option) => (
                    <option key={option} value={option}>
                      {option}
                    </option>
                  ))}
                </select>
                <button className="primary-button" type="submit" disabled={!selectedSchoolId}>
                  Создать
                </button>
              </form>
            </div>

            <div className="panel">
              <div className="panel-head">
                <div>
                  <span className="eyebrow">Список</span>
                  <h3>Школы</h3>
                </div>
              </div>
              <DataTable headers={["Школа", "Статус", "Создана", "Действие"]} emptyMessage="Школ пока нет.">
                {schools.map((item) => (
                  <tr key={item.id}>
                    <td>{item.name}</td>
                    <td><Pill>{item.status}</Pill></td>
                    <td>{formatDateTime(item.created_at)}</td>
                    <td className="table-actions">
                      <button
                        type="button"
                        className="ghost-button"
                        onClick={() => {
                          setSelectedSchoolId(item.id);
                          void withBusy(() => loadSuperAdminData(item.id));
                        }}
                      >
                        Открыть
                      </button>
                      <button
                        type="button"
                        className="icon-button"
                        onClick={() => {
                          if (!window.confirm("Удалить школу?")) return;
                          void handleDelete(() => saas(`/schools/${item.id}`, { method: "DELETE" }), "Школа удалена.");
                        }}
                      >
                        Удалить
                      </button>
                    </td>
                  </tr>
                ))}
              </DataTable>
            </div>

            <div className="two-col">
              <div className="panel">
                <div className="panel-head">
                  <div>
                    <span className="eyebrow">Филиалы</span>
                    <h3>Список филиалов</h3>
                  </div>
                </div>
                <DataTable headers={["Название", "Адрес", "Статус", "Действие"]} emptyMessage="Филиалов нет.">
                  {branches.map((item) => (
                    <tr key={item.id}>
                      <td>{item.name}</td>
                      <td>{item.address || "—"}</td>
                      <td><Pill>{item.status}</Pill></td>
                      <td className="table-actions">
                        <button
                          type="button"
                          className="icon-button"
                          onClick={() => {
                            if (!window.confirm("Удалить филиал?")) return;
                            void handleDelete(() => saas(`/branches/${item.id}`, { method: "DELETE" }), "Филиал удалён.");
                          }}
                        >
                          Удалить
                        </button>
                      </td>
                    </tr>
                  ))}
                </DataTable>
              </div>

              <div className="panel">
                <div className="panel-head">
                  <div>
                    <span className="eyebrow">Пользователи</span>
                    <h3>School-scoped IAM users</h3>
                  </div>
                </div>
                <DataTable headers={["Email", "Роль", "Статус", "Активен"]} emptyMessage="Пользователей нет.">
                  {users.map((item) => (
                    <tr key={item.id}>
                      <td>{item.email}</td>
                      <td><Pill>{item.role}</Pill></td>
                      <td><Pill>{item.status}</Pill></td>
                      <td>{item.is_active ? "Да" : "Нет"}</td>
                    </tr>
                  ))}
                </DataTable>
              </div>
            </div>
          </div>
        </section>
      </>
    );
  }

  function renderOverview() {
    return (
      <>
        <section className="hero-strip">
          <StatCard label="Филиалы" value={branches.length} hint="SaaS branch directory" />
          <StatCard label="Ученики" value={students.length} hint="Student service" />
          <StatCard label="Занятия" value={lessons.length} hint="Lesson scheduling" />
          <StatCard label="Документы" value={documents.length} hint="Metadata + storage" />
        </section>

        <section className="workspace-grid two-col">
          <div className="panel">
            <div className="panel-head">
              <div>
                <span className="eyebrow">Контур школы</span>
                <h3>Сводка по доменам</h3>
              </div>
            </div>
            <div className="detail-grid">
              <Detail label="Автомобили" value={cars.length} />
              <Detail label="Инструкторы" value={instructors.length} />
              <Detail label="Группы" value={groups.length} />
              <Detail label="Платежи" value={payments.length} />
              <Detail label="Шаблоны" value={documentTemplates.length} />
              <Detail label="Роль" value={roleName(me?.role)} />
            </div>
          </div>

          <div className="panel">
            <div className="panel-head">
              <div>
                <span className="eyebrow">Ограничения</span>
                <h3>Права текущего пользователя</h3>
              </div>
            </div>
            <div className="hint-box">
              <strong>{canWriteWorkspace ? "Режим изменения" : "Режим чтения"}</strong>
              <span>
                {canWriteWorkspace
                  ? "Admin может создавать записи в student/payment/document и управлять ресурсами."
                  : "Owner и instructor работают в read-only режиме для school-scoped сервисов."}
              </span>
            </div>
          </div>
        </section>
      </>
    );
  }

  function renderResourcesTab() {
    return (
      <section className="workspace-grid resources-grid">
        <div className="panel tall">
          <div className="panel-head">
            <div>
              <span className="eyebrow">Resources</span>
              <h3>Автомобили и инструкторы</h3>
            </div>
            <div className="head-badges">
              <button
                type="button"
                className={`tab-button ${selectedResourceType === "cars" ? "active" : ""}`}
                onClick={() => {
                  const nextId = cars[0]?.id || "";
                  setSelectedResourceType("cars");
                  setSelectedResourceId(nextId);
                  void loadAvailability("cars", nextId);
                }}
              >
                Авто
              </button>
              <button
                type="button"
                className={`tab-button ${selectedResourceType === "instructors" ? "active" : ""}`}
                onClick={() => {
                  const nextId = instructors[0]?.id || "";
                  setSelectedResourceType("instructors");
                  setSelectedResourceId(nextId);
                  void loadAvailability("instructors", nextId);
                }}
              >
                Инструкторы
              </button>
            </div>
          </div>

          {canWriteWorkspace ? (
            selectedResourceType === "cars" ? (
              <form className="toolbar-form wide" onSubmit={createCar}>
                <BranchSelect branches={branches} name="branch_id" includePlaceholder />
                <select name="instructor_id" defaultValue="">
                  <option value="">Без инструктора</option>
                  {instructors.map((item) => (
                    <option key={item.id} value={item.id}>
                      {item.full_name}
                    </option>
                  ))}
                </select>
                <input name="plate_number" placeholder="Госномер" required />
                <input name="brand" placeholder="Марка" />
                <input name="model" placeholder="Модель" />
                <input name="color" placeholder="Цвет" />
                <select name="transmission" defaultValue="manual">
                  {transmissionTypeOptions.map((option) => (
                    <option key={option} value={option}>
                      {option}
                    </option>
                  ))}
                </select>
                <select name="status" defaultValue="active">
                  {carStatusOptions.map((option) => (
                    <option key={option} value={option}>
                      {option}
                    </option>
                  ))}
                </select>
                <button className="primary-button" type="submit">
                  Добавить авто
                </button>
              </form>
            ) : (
              <form className="toolbar-form wide" onSubmit={createInstructor}>
                <BranchSelect branches={branches} name="branch_id" includePlaceholder />
                <input name="full_name" placeholder="ФИО" required />
                <input name="email" type="email" placeholder="Email" />
                <input name="phone" placeholder="Телефон" />
                <select name="status" defaultValue="active">
                  {instructorStatusOptions.map((option) => (
                    <option key={option} value={option}>
                      {option}
                    </option>
                  ))}
                </select>
                <button className="primary-button" type="submit">
                  Добавить инструктора
                </button>
              </form>
            )
          ) : null}

          <DataTable
            headers={
              selectedResourceType === "cars"
                ? ["Plate", "Branch", "Instructor", "Transmission", "Status", "Availability", "Actions"]
                : ["ФИО", "Филиал", "Email", "Телефон", "Статус", "Действие"]
            }
            emptyMessage="Ресурсы пока не загружены."
          >
            {selectedResourceCollection.map((item) => (
              <tr
                key={item.id}
                className={item.id === selectedResourceId ? "active-row" : ""}
                onClick={() => {
                  setSelectedResourceId(item.id);
                  void loadAvailability(selectedResourceType, item.id);
                }}
              >
                {selectedResourceType === "cars" ? (
                  <>
                    <td>{item.plate_number}</td>
                    <td>{branchNameById[item.branch_id] || item.branch_id}</td>
                    <td>{item.instructor_id ? instructorNameById[item.instructor_id] || item.instructor_id : "—"}</td>
                    <td>{item.transmission || "-"}</td>
                    <td><Pill>{item.status}</Pill></td>
                    <td><Pill>{item.availability_status}</Pill></td>
                    <td className="table-actions">
                      {canWriteWorkspace ? (
                        <button
                          type="button"
                          className="icon-button"
                          onClick={(event) => {
                            event.stopPropagation();
                            if (!window.confirm("Удалить автомобиль?")) return;
                            void handleDelete(
                              () => resources(`/resources/cars/${item.id}`, { method: "DELETE" }),
                              "Автомобиль удалён."
                            );
                          }}
                        >
                          Удалить
                        </button>
                      ) : null}
                    </td>
                  </>
                ) : (
                  <>
                    <td>{item.full_name}</td>
                    <td>{branchNameById[item.branch_id] || item.branch_id}</td>
                    <td>{item.email || "—"}</td>
                    <td>{item.phone || "—"}</td>
                    <td><Pill>{item.status}</Pill></td>
                    <td className="table-actions">
                      {canWriteWorkspace ? (
                        <button
                          type="button"
                          className="icon-button"
                          onClick={(event) => {
                            event.stopPropagation();
                            if (!window.confirm("Удалить инструктора?")) return;
                            void handleDelete(
                              () => resources(`/resources/instructors/${item.id}`, { method: "DELETE" }),
                              "Инструктор удалён."
                            );
                          }}
                        >
                          Удалить
                        </button>
                      ) : null}
                    </td>
                  </>
                )}
              </tr>
            ))}
          </DataTable>
        </div>

        <div className="stack-column">
          <div className="panel">
            <div className="panel-head">
              <div>
                <span className="eyebrow">Карточка</span>
                <h3>Выбранный ресурс</h3>
              </div>
            </div>
            {selectedResource ? (
              <>
                <div className="detail-grid">
                  <Detail label="ID" value={selectedResource.id} />
                  <Detail label="Branch" value={branchNameById[selectedResource.branch_id] || selectedResource.branch_id} />
                  {selectedResourceType === "cars" ? (
                    <Detail label="Transmission" value={selectedResource.transmission || "-"} />
                  ) : null}
                  <Detail label="Status" value={selectedResource.status} />
                  <Detail label="Updated" value={formatDateTime(selectedResource.updated_at)} />
                </div>
                {canWriteWorkspace ? (
                  <form key={`${selectedResourceType}-${selectedResource.id}`} className="toolbar-form wide" onSubmit={updateSelectedResource}>
                    <select name="branch_id" defaultValue={selectedResource.branch_id}>
                      {branches.map((branch) => (
                        <option key={branch.id} value={branch.id}>
                          {branch.name}
                        </option>
                      ))}
                    </select>
                    {selectedResourceType === "cars" ? (
                      <>
                        <select name="instructor_id" defaultValue={selectedResource.instructor_id || ""}>
                          <option value="">No instructor</option>
                          {instructors.map((item) => (
                            <option key={item.id} value={item.id}>
                              {item.full_name}
                            </option>
                          ))}
                        </select>
                        <input name="plate_number" defaultValue={selectedResource.plate_number || ""} placeholder="Plate" required />
                        <input name="brand" defaultValue={selectedResource.brand || ""} placeholder="Brand" />
                        <input name="model" defaultValue={selectedResource.model || ""} placeholder="Model" />
                        <input name="color" defaultValue={selectedResource.color || ""} placeholder="Color" />
                        <select name="transmission" defaultValue={selectedResource.transmission || "manual"}>
                          {transmissionTypeOptions.map((option) => (
                            <option key={option} value={option}>
                              {option}
                            </option>
                          ))}
                        </select>
                        <select name="status" defaultValue={selectedResource.status || "active"}>
                          {carStatusOptions.map((option) => (
                            <option key={option} value={option}>
                              {option}
                            </option>
                          ))}
                        </select>
                      </>
                    ) : (
                      <>
                        <input name="full_name" defaultValue={selectedResource.full_name || ""} placeholder="Full name" required />
                        <input name="email" type="email" defaultValue={selectedResource.email || ""} placeholder="Email" />
                        <input name="phone" defaultValue={selectedResource.phone || ""} placeholder="Phone" />
                        <select name="status" defaultValue={selectedResource.status || "active"}>
                          {instructorStatusOptions.map((option) => (
                            <option key={option} value={option}>
                              {option}
                            </option>
                          ))}
                        </select>
                      </>
                    )}
                    <button className="primary-button" type="submit">
                      Save resource
                    </button>
                  </form>
                ) : null}
              </>
            ) : (
              <EmptyState message="Выберите ресурс в таблице." />
            )}
          </div>

          <div className="panel">
            <div className="panel-head">
              <div>
                <span className="eyebrow">Availability</span>
                <h3>Окна доступности</h3>
              </div>
            </div>

            {selectedResource ? (
              <>
                {canWriteWorkspace ? (
                  <form className="recurring-availability-form" onSubmit={addRecurringAvailability}>
                    <div className="form-section-title">
                      <strong>Повторяющееся расписание</strong>
                    </div>
                    <div className="recurring-availability-grid">
                      <label>
                        Тип
                        <select name="type" defaultValue="working_hours">
                          {availabilityTypes.map((option) => (
                            <option key={option} value={option}>
                              {option}
                            </option>
                          ))}
                        </select>
                      </label>
                      <label>
                        Начать с
                        <input name="start_date" type="date" defaultValue={todayDateInput()} required />
                      </label>
                      <label>
                        Месяцев
                        <input name="repeat_months" type="number" min="1" max="12" defaultValue="3" required />
                      </label>
                      <label>
                        С
                        <input name="start_time" type="time" defaultValue="09:00" required />
                      </label>
                      <label>
                        До
                        <input name="end_time" type="time" defaultValue="18:00" required />
                      </label>
                      <label className="recurring-comment">
                        Комментарий
                        <input name="comment" placeholder="Например, стандартная неделя" />
                      </label>
                    </div>
                    <div className="weekday-picker">
                      {recurringWeekdayOptions.map(([value, label]) => (
                        <label key={value} className="weekday-chip">
                          <input name={`weekday_${value}`} type="checkbox" defaultChecked={["monday", "tuesday", "wednesday", "thursday", "friday"].includes(value)} />
                          <span>{label}</span>
                        </label>
                      ))}
                    </div>
                    <button className="ghost-button" type="submit">
                      Добавить повторяющиеся окна
                    </button>
                  </form>
                ) : null}

                <div className="availability-list">
                  {availabilityDrafts.length ? (
                    availabilityDrafts.map((item) => (
                      <div key={item.id} className="availability-row">
                        <select
                          value={item.type}
                          onChange={(event) => updateAvailabilityWindow(item.id, "type", event.target.value)}
                          disabled={!canWriteWorkspace}
                        >
                          {availabilityTypes.map((option) => (
                            <option key={option} value={option}>
                              {option}
                            </option>
                          ))}
                        </select>
                        <input
                          type="datetime-local"
                          value={item.start_at}
                          onChange={(event) => updateAvailabilityWindow(item.id, "start_at", event.target.value)}
                          disabled={!canWriteWorkspace}
                        />
                        <input
                          type="datetime-local"
                          value={item.end_at}
                          onChange={(event) => updateAvailabilityWindow(item.id, "end_at", event.target.value)}
                          disabled={!canWriteWorkspace}
                        />
                        <input
                          value={item.comment}
                          placeholder="Комментарий"
                          onChange={(event) => updateAvailabilityWindow(item.id, "comment", event.target.value)}
                          disabled={!canWriteWorkspace}
                        />
                        {canWriteWorkspace ? (
                          <button type="button" className="icon-button" onClick={() => removeAvailabilityWindow(item.id)}>
                            Удалить
                          </button>
                        ) : null}
                      </div>
                    ))
                  ) : (
                    <EmptyState compact message="Окон пока нет." />
                  )}
                </div>

                {canWriteWorkspace ? (
                  <div className="head-badges">
                    <button type="button" className="ghost-button" onClick={addAvailabilityWindow}>
                      Добавить окно
                    </button>
                    <button type="button" className="primary-button" onClick={saveAvailability} disabled={!selectedResourceId}>
                      Сохранить availability
                    </button>
                  </div>
                ) : null}
              </>
            ) : (
              <EmptyState message="Сначала выберите ресурс." />
            )}
          </div>

        </div>
      </section>
    );
  }

  function renderStudentsTab() {
    return (
      <section className="workspace-grid two-col">
        <div className="stack-column">
          {canWriteWorkspace ? (
            <div className="panel">
              <div className="panel-head">
                <div>
                  <span className="eyebrow">Students</span>
                  <h3>Создать ученика</h3>
                </div>
              </div>
              <form className="toolbar-form wide" onSubmit={createStudent}>
                <BranchSelect branches={branches} name="branch_id" includePlaceholder />
                <input name="full_name" placeholder="ФИО" required />
                <input name="phone" placeholder="Телефон" required />
                <select name="status" defaultValue="active">
                  {studentStatusOptions.map((option) => (
                    <option key={option} value={option}>
                      {option}
                    </option>
                  ))}
                </select>
                <button className="primary-button" type="submit">
                  Создать
                </button>
              </form>
            </div>
          ) : null}

          <div className="panel">
            <div className="panel-head">
              <div>
                <span className="eyebrow">Students</span>
                <h3>Список учеников</h3>
              </div>
            </div>
            <DataTable headers={["ФИО", "Телефон", "Филиал", "Статус", "Действие"]} emptyMessage="Учеников нет.">
              {students.map((item) => (
                <tr key={item.id}>
                  <td>{item.full_name}</td>
                  <td>{item.phone}</td>
                  <td>{branchNameById[item.branch_id] || item.branch_id}</td>
                  <td><Pill>{item.status}</Pill></td>
                  <td className="table-actions">
                    <button type="button" className="ghost-button" onClick={() => setSelectedStudentId(item.id)}>
                      Выбрать
                    </button>
                    {canWriteWorkspace ? (
                      <button
                        type="button"
                        className="icon-button"
                        onClick={() => {
                          if (!window.confirm("Удалить ученика?")) return;
                          void handleDelete(() => studentsApi(`/students/${item.id}`, { method: "DELETE" }), "Ученик удалён.");
                        }}
                      >
                        Удалить
                      </button>
                    ) : null}
                  </td>
                </tr>
              ))}
            </DataTable>
          </div>
        </div>

        <div className="stack-column">
          {canWriteWorkspace ? (
            <div className="panel">
              <div className="panel-head">
                <div>
                  <span className="eyebrow">Programs</span>
                  <h3>Create Study Program</h3>
                </div>
              </div>
              <form className="toolbar-form" onSubmit={createStudyProgram}>
                <input name="name" placeholder="Program name" required />
                <select name="transmission" defaultValue="manual" required>
                  {transmissionTypeOptions.map((option) => (
                    <option key={option} value={option}>
                      {option}
                    </option>
                  ))}
                </select>
                <input name="total_hours" type="number" min="0" step="0.5" placeholder="Total hours" required />
                <button className="primary-button" type="submit">
                  Create
                </button>
              </form>
            </div>
          ) : null}

          {canWriteWorkspace && editingProgram ? (
            <div className="panel">
              <div className="panel-head">
                <div>
                  <span className="eyebrow">Programs</span>
                  <h3>Edit Study Program</h3>
                </div>
                <button type="button" className="ghost-button" onClick={() => setEditingProgramId("")}>
                  Hide
                </button>
              </div>
              <form key={editingProgram.id} className="toolbar-form" onSubmit={updateStudyProgram}>
                <input name="name" defaultValue={editingProgram.name || ""} placeholder="Program name" required />
                <select name="transmission" defaultValue={editingProgram.transmission || "manual"} required>
                  {transmissionTypeOptions.map((option) => (
                    <option key={option} value={option}>
                      {option}
                    </option>
                  ))}
                </select>
                <input
                  name="total_hours"
                  type="number"
                  min="0"
                  step="0.5"
                  defaultValue={editingProgram.total_hours ?? 0}
                  placeholder="Total hours"
                  required
                />
                <button className="primary-button" type="submit">
                  Save
                </button>
              </form>
            </div>
          ) : null}

          <div className="panel">
            <div className="panel-head">
              <div>
                <span className="eyebrow">Programs</span>
                <h3>Study Program List</h3>
              </div>
            </div>
            <DataTable headers={["Name", "Transmission", "Total hours", "Created", "Actions"]} emptyMessage="No study programs yet.">
              {studyPrograms.map((item) => (
                <tr key={item.id}>
                  <td>{item.name}</td>
                  <td>{item.transmission}</td>
                  <td>{item.total_hours}</td>
                  <td>{formatDateTime(item.created_at)}</td>
                  <td className="table-actions">
                    {canWriteWorkspace ? (
                      <button type="button" className="ghost-button" onClick={() => setEditingProgramId(item.id)}>
                        Edit
                      </button>
                    ) : null}
                    {canWriteWorkspace ? (
                      <button
                        type="button"
                        className="icon-button"
                        onClick={() => {
                          if (!window.confirm("Delete study program?")) return;
                          void handleDelete(() => studentsApi(`/study-programs/${item.id}`, { method: "DELETE" }), "Study program deleted.");
                        }}
                      >
                        Delete
                      </button>
                    ) : null}
                  </td>
                </tr>
              ))}
            </DataTable>
          </div>

          {canWriteWorkspace ? (
            <div className="panel">
              <div className="panel-head">
                <div>
                  <span className="eyebrow">Groups</span>
                  <h3>Создать группу</h3>
                </div>
              </div>
              <form className="toolbar-form" onSubmit={createGroup}>
                <BranchSelect branches={branches} name="branch_id" includePlaceholder />
                <input name="name" placeholder="Название группы" required />
                <button className="primary-button" type="submit">
                  Создать
                </button>
              </form>
            </div>
          ) : null}

          <div className="panel">
            <div className="panel-head">
              <div>
                <span className="eyebrow">Groups</span>
                <h3>Список групп</h3>
              </div>
            </div>
            <DataTable headers={["Название", "Филиал", "Создана", "Действие"]} emptyMessage="Групп нет.">
              {groups.map((item) => (
                <tr key={item.id}>
                  <td>{item.name}</td>
                  <td>{branchNameById[item.branch_id] || item.branch_id}</td>
                  <td>{formatDateTime(item.created_at)}</td>
                  <td className="table-actions">
                    {canWriteWorkspace ? (
                      <button
                        type="button"
                        className="icon-button"
                        onClick={() => {
                          if (!window.confirm("Удалить группу?")) return;
                          void handleDelete(() => studentsApi(`/groups/${item.id}`, { method: "DELETE" }), "Группа удалена.");
                        }}
                      >
                        Удалить
                      </button>
                    ) : null}
                  </td>
                </tr>
              ))}
            </DataTable>
          </div>
        </div>
      </section>
    );
  }

  function renderLessonsTab() {
    return (
      <section className="stack-column">
        {canWriteWorkspace ? (
          <div className="panel">
            <div className="panel-head">
              <div>
                <span className="eyebrow">Lessons</span>
                <h3>Создать занятие</h3>
              </div>
            </div>
            <form className="toolbar-form wide" onSubmit={createLesson}>
              <BranchSelect branches={branches} name="branch_id" includePlaceholder />
              <select name="student_id" defaultValue="">
                <option value="">Ученик</option>
                {students.map((item) => (
                  <option key={item.id} value={item.id}>
                    {item.full_name} - {groupNameById[item.group_id] || "No group"}
                  </option>
                ))}
              </select>
              <select name="instructor_id" defaultValue="">
                <option value="">Инструктор</option>
                {instructors.map((item) => (
                  <option key={item.id} value={item.id}>
                    {item.full_name} - {groupNameById[item.group_id] || "No group"}
                  </option>
                ))}
              </select>
              <select name="car_id" defaultValue="">
                <option value="">Авто</option>
                {cars.map((item) => (
                  <option key={item.id} value={item.id}>
                    {carNameById[item.id]}
                  </option>
                ))}
              </select>
              <input name="start_at" type="datetime-local" defaultValue={nextDateTime(1)} />
              <input name="end_at" type="datetime-local" defaultValue={nextDateTime(3)} />
              <select name="status" defaultValue="scheduled">
                {lessonStatusOptions.map((option) => (
                  <option key={option} value={option}>
                    {option}
                  </option>
                ))}
              </select>
              <input name="comment" placeholder="Комментарий" />
              <button className="primary-button" type="submit">
                Создать
              </button>
            </form>
          </div>
        ) : null}

        <div className="panel">
          <div className="panel-head">
            <div>
              <span className="eyebrow">Lessons</span>
              <h3>Список занятий</h3>
            </div>
          </div>
          <DataTable
            headers={["Ученик", "Инструктор", "Авто", "Начало", "Окончание", "Статус", "Действие"]}
            emptyMessage="Занятий нет."
          >
            {lessons.map((item) => (
              <tr key={item.id}>
                <td>{studentNameById[item.student_id] || item.student_id}</td>
                <td>{instructorNameById[item.instructor_id] || item.instructor_id}</td>
                <td>{carNameById[item.car_id] || item.car_id}</td>
                <td>{formatDateTime(item.start_at)}</td>
                <td>{formatDateTime(item.end_at)}</td>
                <td><Pill>{item.status}</Pill></td>
                <td className="table-actions">
                  {canWriteWorkspace ? (
                    <button
                      type="button"
                      className="icon-button"
                      onClick={() => {
                        if (!window.confirm("Удалить занятие?")) return;
                        void handleDelete(() => lessonsApi(`/lessons/${item.id}`, { method: "DELETE" }), "Занятие удалено.");
                      }}
                    >
                      Удалить
                    </button>
                  ) : null}
                </td>
              </tr>
            ))}
          </DataTable>
        </div>
      </section>
    );
  }

  function renderPaymentsTab() {
    return (
      <section className="workspace-grid two-col">
        <div className="stack-column">
          <div className="panel">
            <div className="panel-head">
              <div>
                <span className="eyebrow">Student Focus</span>
                <h3>Баланс ученика</h3>
              </div>
              <select value={selectedStudentId} onChange={(event) => setSelectedStudentId(event.target.value)}>
                <option value="">Выберите ученика</option>
                {students.map((item) => (
                  <option key={item.id} value={item.id}>
                    {item.full_name} - {groupNameById[item.group_id] || "No group"}
                  </option>
                ))}
              </select>
            </div>

            {studentBalance ? (
              <div className="detail-grid">
                <Detail label="Договор" value={formatMoney(studentBalance.contract_total)} />
                <Detail label="Оплачено" value={formatMoney(studentBalance.paid_amount)} />
                <Detail label="Остаток" value={formatMoney(studentBalance.remaining_amount)} />
                <Detail label="Student ID" value={studentBalance.student_id} />
              </div>
            ) : (
              <EmptyState message="Выберите ученика, чтобы увидеть баланс." />
            )}
          </div>

          {canWriteWorkspace ? (
            <>
              <div className="panel">
                <div className="panel-head">
                  <div>
                    <span className="eyebrow">Payments</span>
                    <h3>Создать платёж</h3>
                  </div>
                </div>
                <form className="toolbar-form wide" onSubmit={createPayment}>
                  <select name="student_id" defaultValue={selectedStudentId || ""}>
                    <option value="">Ученик</option>
                    {students.map((item) => (
                      <option key={item.id} value={item.id}>
                        {item.full_name}
                      </option>
                    ))}
                  </select>
                  <select name="type" defaultValue="cash">
                    {paymentTypeOptions.map((option) => (
                      <option key={option} value={option}>
                        {option}
                      </option>
                    ))}
                  </select>
                  <input name="amount" type="number" min="0.01" step="0.01" placeholder="Сумма" required />
                  <input name="comment" placeholder="Комментарий" />
                  <button className="primary-button" type="submit">
                    Создать
                  </button>
                </form>
              </div>

              <div className="panel">
                <div className="panel-head">
                  <div>
                    <span className="eyebrow">Contract</span>
                    <h3>Обновить сумму договора</h3>
                  </div>
                </div>
                <form className="toolbar-form" onSubmit={updateContract}>
                  <select name="student_id" defaultValue={selectedStudentId || ""}>
                    <option value="">Ученик</option>
                    {students.map((item) => (
                      <option key={item.id} value={item.id}>
                        {item.full_name}
                      </option>
                    ))}
                  </select>
                  <input name="contract_total" type="number" min="0" step="0.01" placeholder="Сумма договора" required />
                  <button className="primary-button" type="submit">
                    Обновить
                  </button>
                </form>
              </div>
            </>
          ) : null}
        </div>

        <div className="panel">
          <div className="panel-head">
            <div>
              <span className="eyebrow">Payments</span>
              <h3>Журнал платежей</h3>
            </div>
          </div>
          <DataTable headers={["Ученик", "Тип", "Сумма", "Комментарий", "Создан", "Действие"]} emptyMessage="Платежей нет.">
            {payments.map((item) => (
              <tr key={item.id}>
                <td>{studentNameById[item.student_id] || item.student_id}</td>
                <td><Pill>{item.type}</Pill></td>
                <td>{formatMoney(item.amount)}</td>
                <td>{item.comment || "—"}</td>
                <td>{formatDateTime(item.created_at)}</td>
                <td className="table-actions">
                  {canWriteWorkspace ? (
                    <button
                      type="button"
                      className="icon-button"
                      onClick={() => {
                        if (!window.confirm("Удалить платёж?")) return;
                        void handleDelete(() => paymentsApi(`/payments/${item.id}`, { method: "DELETE" }), "Платёж удалён.");
                      }}
                    >
                      Удалить
                    </button>
                  ) : null}
                </td>
              </tr>
            ))}
          </DataTable>
        </div>
      </section>
    );
  }

  function renderDocumentsTab() {
    return (
      <section className="stack-column">
        <div className="workspace-grid two-col">
          <div className="stack-column">
            {canWriteWorkspace ? (
              <div className="panel">
                <div className="panel-head">
                  <div>
                    <span className="eyebrow">Upload</span>
                    <h3>Загрузить документ</h3>
                  </div>
                </div>
                <form className="form-stack" onSubmit={uploadDocument}>
                  <select name="student_id" defaultValue={selectedStudentId || ""}>
                    <option value="">Ученик</option>
                    {students.map((item) => (
                      <option key={item.id} value={item.id}>
                        {item.full_name}
                      </option>
                    ))}
                  </select>
                  <input name="type" placeholder="Тип документа" required />
                  <input name="title" placeholder="Заголовок" required />
                  <input name="file" type="file" required />
                  <button className="primary-button" type="submit">
                    Загрузить
                  </button>
                </form>
              </div>
            ) : null}

            {canWriteWorkspace ? (
              <div className="panel">
                <div className="panel-head">
                  <div>
                    <span className="eyebrow">Templates</span>
                    <h3>Создать шаблон</h3>
                  </div>
                </div>
                <form className="form-stack" onSubmit={createTemplate}>
                  <input name="name" placeholder="Название шаблона" required />
                  <input name="type" placeholder="Тип документа" required />
                  <textarea name="content" rows="8" placeholder="Текст с placeholder'ами: {{.full_name}}" required />
                  <button className="primary-button" type="submit">
                    Создать шаблон
                  </button>
                </form>
              </div>
            ) : null}
          </div>

          <div className="stack-column">
            {canWriteWorkspace ? (
              <div className="panel">
                <div className="panel-head">
                  <div>
                    <span className="eyebrow">Generate</span>
                    <h3>Сгенерировать документ</h3>
                  </div>
                </div>
                <form className="form-stack" onSubmit={generateDocument}>
                  <select name="student_id" defaultValue={selectedStudentId || ""}>
                    <option value="">Ученик</option>
                    {students.map((item) => (
                      <option key={item.id} value={item.id}>
                        {item.full_name}
                      </option>
                    ))}
                  </select>
                  <select name="template_id" defaultValue="">
                    <option value="">Шаблон</option>
                    {documentTemplates.map((item) => (
                      <option key={item.id} value={item.id}>
                        {item.name}
                      </option>
                    ))}
                  </select>
                  <input name="title" placeholder="Заголовок документа" required />
                  <input name="file_name" placeholder="Имя файла (опционально)" />
                  <textarea
                    name="placeholders"
                    rows="7"
                    placeholder={"full_name=Иван Иванов\ncontract_total=50000"}
                  />
                  <button className="primary-button" type="submit">
                    Сгенерировать
                  </button>
                </form>
              </div>
            ) : null}

            <div className="panel">
              <div className="panel-head">
                <div>
                  <span className="eyebrow">Templates</span>
                  <h3>Шаблоны документов</h3>
                </div>
              </div>
              <DataTable headers={["Название", "Тип", "Создан", "Действие"]} emptyMessage="Шаблонов нет.">
                {documentTemplates.map((item) => (
                  <tr key={item.id}>
                    <td>{item.name}</td>
                    <td><Pill>{item.type}</Pill></td>
                    <td>{formatDateTime(item.created_at)}</td>
                    <td className="table-actions">
                      {canWriteWorkspace ? (
                        <button
                          type="button"
                          className="icon-button"
                          onClick={() => {
                            if (!window.confirm("Удалить шаблон?")) return;
                            void handleDelete(
                              () => documentsApi(`/document-templates/${item.id}`, { method: "DELETE" }),
                              "Шаблон удалён."
                            );
                          }}
                        >
                          Удалить
                        </button>
                      ) : null}
                    </td>
                  </tr>
                ))}
              </DataTable>
            </div>
          </div>
        </div>

        <div className="panel">
          <div className="panel-head">
            <div>
              <span className="eyebrow">Documents</span>
              <h3>Реестр документов</h3>
            </div>
          </div>
          <DataTable
            headers={["Заголовок", "Ученик", "Тип", "Статус", "Файл", "Размер", "Действие"]}
            emptyMessage="Документов нет."
          >
            {documents.map((item) => (
              <tr key={item.id}>
                <td>{item.title}</td>
                <td>{studentNameById[item.student_id] || item.student_id}</td>
                <td><Pill>{item.type}</Pill></td>
                <td><Pill>{item.status}</Pill></td>
                <td>{item.mime_type}</td>
                <td>{item.size} bytes</td>
                <td className="table-actions">
                  <button type="button" className="ghost-button" onClick={() => void downloadDocument(item.id)}>
                    Скачать
                  </button>
                  {canWriteWorkspace ? (
                    <button
                      type="button"
                      className="icon-button"
                      onClick={() => {
                        if (!window.confirm("Удалить документ?")) return;
                        void handleDelete(() => documentsApi(`/documents/${item.id}`, { method: "DELETE" }), "Документ удалён.");
                      }}
                    >
                      Удалить
                    </button>
                  ) : null}
                </td>
              </tr>
            ))}
          </DataTable>
        </div>
      </section>
    );
  }

  function renderPlanningTab() {
    const planningGroups = groups.map((group) => ({
      ...group,
      students: students.filter((student) => student.group_id === group.id),
    }));
    const ungroupedPlanningStudents = students.filter((student) => !student.group_id);

    return (
      <section className="workspace-grid planning-grid">
        <div className="panel planning-panel">
          <div className="panel-head">
            <div>
              <span className="eyebrow">Planning</span>
              <h3>Build schedule</h3>
              <p className="muted-text">Select the time range, students or groups, branches, and how the generated lessons should be applied.</p>
            </div>
          </div>
          {canWriteWorkspace ? (
            <form className="planning-form" onSubmit={requestPlanning}>
              <div className="planning-section">
                <div>
                  <span className="eyebrow">Step 1</span>
                  <h4>Planning period</h4>
                </div>
                <div className="planning-fields two">
                  <label>
                    Start date and time
                    <input name="period_start" type="datetime-local" defaultValue={nextDateTime(24)} required />
                    <span>First slot where planning can place lessons.</span>
                  </label>
                  <label>
                    End date and time
                    <input name="period_end" type="datetime-local" defaultValue={nextDateTime(120)} required />
                    <span>Last moment before which lessons can be created.</span>
                  </label>
                </div>
              </div>

              <div className="planning-section">
                <div>
                  <span className="eyebrow">Step 2</span>
                  <h4>Students and groups</h4>
                </div>
                <div className="planning-fields two">
                  <label>
                    Individual students
                    <select name="student_ids" multiple size="10">
                      {students.map((item) => (
                        <option key={item.id} value={item.id}>
                          {item.full_name} - {branchNameById[item.branch_id] || "branch not set"}
                        </option>
                      ))}
                    </select>
                    <span>Select exact students. Hold Ctrl/Cmd for multiple selection.</span>
                  </label>
                  <label>
                    Groups
                    <select name="group_ids" multiple size="10">
                      {groups.map((item) => (
                        <option key={item.id} value={item.id}>
                          {item.name} - {branchNameById[item.branch_id] || "branch not set"}
                        </option>
                      ))}
                    </select>
                    <span>Planning receives selected groups. The preview on the right shows students assigned to each group.</span>
                  </label>
                </div>
              </div>

              <div className="planning-section">
                <div>
                  <span className="eyebrow">Step 3</span>
                  <h4>Scope and apply mode</h4>
                </div>
                <div className="planning-fields three">
                  <label>
                    Branches
                    <select name="branch_ids" multiple size="5">
                      {branches.map((item) => (
                        <option key={item.id} value={item.id}>
                          {item.name}
                        </option>
                      ))}
                    </select>
                    <span>Use this to limit students and resources to selected branches.</span>
                  </label>
                  <label>
                    Apply strategy
                    <select name="replace_strategy" defaultValue="append">
                      {planningReplaceStrategies.map((option) => (
                        <option key={option} value={option}>
                          {option}
                        </option>
                      ))}
                    </select>
                    <span>append adds lessons, merge combines, replace_generated_only replaces only generated lessons.</span>
                  </label>
                  <label>
                    Request ID
                    <input name="request_id" placeholder="Optional, e.g. plan-week-18" />
                    <span>Leave empty unless you need a custom id for tracing.</span>
                  </label>
                </div>
              </div>

              <button className="primary-button planning-submit" type="submit">
                Send Planning request
              </button>
            </form>
          ) : (
            <EmptyState message="Only admin can enqueue Planning requests." />
          )}
        </div>

        <div className="stack-column">
          <div className="panel">
            <div className="panel-head">
              <div>
                <span className="eyebrow">Groups</span>
                <h3>Students by group</h3>
                <p className="muted-text">Students are shown by their assigned group_id.</p>
              </div>
            </div>
            <div className="planning-group-list">
              {planningGroups.length ? (
                planningGroups.map((group) => (
                  <div key={group.id} className="planning-group-card">
                    <div className="entity-card-head">
                      <div>
                        <strong>{group.name}</strong>
                        <div className="entity-meta">{branchNameById[group.branch_id] || group.branch_id}</div>
                      </div>
                      <Pill>{group.students.length} students</Pill>
                    </div>
                    <div className="planning-student-list">
                      {group.students.length ? (
                        group.students.map((student) => (
                          <span key={student.id} className="planning-student-chip">
                            {student.full_name}
                          </span>
                        ))
                      ) : (
                        <span className="muted-text">No students in this group's branch.</span>
                      )}
                    </div>
                  </div>
                ))
              ) : (
                <EmptyState compact message="No groups yet." />
              )}
              {ungroupedPlanningStudents.length ? (
                <div className="planning-group-card">
                  <div className="entity-card-head">
                    <div>
                      <strong>Without group</strong>
                      <div className="entity-meta">Students without group_id</div>
                    </div>
                    <Pill>{ungroupedPlanningStudents.length} students</Pill>
                  </div>
                  <div className="planning-student-list">
                    {ungroupedPlanningStudents.map((student) => (
                      <span key={student.id} className="planning-student-chip">
                        {student.full_name}
                      </span>
                    ))}
                  </div>
                </div>
              ) : null}
            </div>
          </div>

          <div className="panel">
            <div className="panel-head">
              <div>
                <span className="eyebrow">Contract</span>
                <h3>Kafka payload</h3>
              </div>
            </div>
            <div className="hint-box">
              <strong>Planning Service does not save lessons directly</strong>
              <span>Lesson service accepts this HTTP request, validates tenant context, and publishes a Kafka task.</span>
            </div>
            <pre className="code-box">{`{
  "period_start": "datetime",
  "period_end": "datetime",
  "student_ids": ["uuid"],
  "group_ids": ["uuid"],
  "branch_ids": ["uuid"],
  "replace_strategy": "append|merge|replace_generated_only",
  "request_id": "string"
}`}</pre>
          </div>
        </div>
      </section>
    );
  }


  function renderReportingTab() {
    return (
      <section className="stack-column">
        <section className="hero-strip">
          <StatCard
            label="Активные ученики"
            value={reportingDashboard ? reportingDashboard.active_students : students.length}
            hint="Агрегат reporting service"
          />
          <StatCard
            label="Занятия"
            value={reportingDashboard ? reportingDashboard.lessons_total : lessons.length}
            hint="Общее количество занятий"
          />
          <StatCard
            label="Поступления"
            value={formatMoney(reportingDashboard?.payments_total)}
            hint="Сумма платежей по школе"
          />
          <StatCard
            label="Остаток"
            value={formatMoney(reportingDashboard?.contracts_remaining_total)}
            hint="Остаток по договорам"
          />
        </section>

        <div className="workspace-grid two-col">
          <div className="panel">
            <div className="panel-head">
              <div>
                <span className="eyebrow">Reporting</span>
                <h3>Snapshot</h3>
              </div>
              <button className="ghost-button" type="button" onClick={() => void withBusy(() => loadReportingData())}>
                Refresh
              </button>
            </div>
            <form className="toolbar-form wide" onSubmit={applyReportingFilters}>
              <select name="branch_id" defaultValue={reportingFilters.branch_id}>
                <option value="">All branches</option>
                {branches.map((branch) => (
                  <option key={branch.id} value={branch.id}>
                    {branch.name}
                  </option>
                ))}
              </select>
              <select name="group_id" defaultValue={reportingFilters.group_id}>
                <option value="">All groups</option>
                {groups.map((group) => (
                  <option key={group.id} value={group.id}>
                    {group.name}
                  </option>
                ))}
              </select>
              <input name="month" type="month" defaultValue={reportingFilters.month} />
              <button className="primary-button" type="submit">
                Build report
              </button>
            </form>
            <div className="detail-grid">
              <Detail label="Students Total" value={reportingStudents?.total} />
              <Detail label="Students Active" value={reportingStudents?.active} />
              <Detail label="Students New" value={reportingStudents?.new} />
              <Detail label="Lessons Scheduled" value={reportingLessons?.scheduled} />
              <Detail label="Lessons Completed" value={reportingLessons?.completed} />
              <Detail label="Lessons Cancelled" value={reportingLessons?.cancelled} />
              <Detail label="Instructors" value={reportingDashboard?.instructors_total} />
              <Detail label="Cars" value={reportingDashboard?.cars_total} />
            </div>
          </div>

          <div className="panel">
            <div className="panel-head">
              <div>
                <span className="eyebrow">Finance</span>
                <h3>Payment Breakdown</h3>
              </div>
            </div>
            <DataTable headers={["Type", "Amount"]} emptyMessage="No payment aggregates yet.">
              {(reportingFinance?.payment_types || []).map((item) => (
                <tr key={item.type}>
                  <td><Pill>{item.type}</Pill></td>
                  <td>{formatMoney(item.amount)}</td>
                </tr>
              ))}
            </DataTable>
          </div>
        </div>

        <div className="workspace-grid two-col">
          <div className="panel">
            <div className="panel-head">
              <div>
                <span className="eyebrow">Workload</span>
                <h3>Instructor Load</h3>
              </div>
            </div>
            <DataTable headers={["Instructor", "Lessons", "Minutes"]} emptyMessage="No instructor load yet.">
              {reportingInstructorWorkload.map((item) => (
                <tr key={item.instructor_id}>
                  <td>{item.full_name}</td>
                  <td>{item.lessons_total}</td>
                  <td>{item.total_minutes}</td>
                </tr>
              ))}
            </DataTable>
          </div>

          <div className="panel">
            <div className="panel-head">
              <div>
                <span className="eyebrow">Utilization</span>
                <h3>Cars Load</h3>
              </div>
            </div>
            <DataTable headers={["Car", "Plate", "Lessons", "Minutes"]} emptyMessage="No car utilization yet.">
              {reportingCarsUtilization.map((item) => (
                <tr key={item.car_id}>
                  <td>{item.display_name}</td>
                  <td>{item.plate_number}</td>
                  <td>{item.lessons_total}</td>
                  <td>{item.total_minutes}</td>
                </tr>
              ))}
            </DataTable>
          </div>
        </div>
      </section>
    );
  }

  function renderStudentsTabV2() {
    return (
      <section className="workspace-grid two-col">
        <div className="stack-column">
          {canWriteWorkspace ? (
            <div className="panel">
              <div className="panel-head">
                <div>
                  <span className="eyebrow">Students</span>
                  <h3>Create Student</h3>
                </div>
              </div>
              <form className="toolbar-form wide" onSubmit={createStudent}>
                <BranchSelect branches={branches} name="branch_id" includePlaceholder />
                <select name="group_id" defaultValue="">
                  <option value="">No group</option>
                  {groups.map((group) => (
                    <option key={group.id} value={group.id}>
                      {group.name}
                    </option>
                  ))}
                </select>
                <select name="program_id" defaultValue="" required>
                  <option value="">Program</option>
                  {studyPrograms.map((program) => (
                    <option key={program.id} value={program.id}>
                      {program.name}
                    </option>
                  ))}
                </select>
                <input name="full_name" placeholder="Full name" required />
                <input name="phone" placeholder="Phone" required />
                <select name="status" defaultValue="active">
                  {studentStatusOptions.map((option) => (
                    <option key={option} value={option}>
                      {option}
                    </option>
                  ))}
                </select>
                <select name="preferred_weekdays" multiple size="4">
                  {studentPreferredWeekdayOptions.map((option) => (
                    <option key={option} value={option}>
                      {option}
                    </option>
                  ))}
                </select>
                <select name="preferred_time_windows" multiple size="3">
                  {studentPreferredTimeWindowOptions.map((option) => (
                    <option key={option} value={option}>
                      {option}
                    </option>
                  ))}
                </select>
                <input name="manual_hours_completed" type="number" min="0" step="0.5" defaultValue="0" placeholder="Manual hours completed" />
                <button className="primary-button" type="submit">
                  Create
                </button>
              </form>
            </div>
          ) : null}

          {canWriteWorkspace && editingStudent ? (
            <div className="panel">
              <div className="panel-head">
                <div>
                  <span className="eyebrow">Students</span>
                  <h3>Edit Student</h3>
                </div>
                <button type="button" className="ghost-button" onClick={() => setEditingStudentId("")}>
                  Hide
                </button>
              </div>
              <form key={editingStudent.id} className="toolbar-form wide" onSubmit={updateStudent}>
                <select name="branch_id" defaultValue={editingStudent.branch_id}>
                  {branches.map((branch) => (
                    <option key={branch.id} value={branch.id}>
                      {branch.name}
                    </option>
                  ))}
                </select>
                <select name="group_id" defaultValue={editingStudent.group_id || ""}>
                  <option value="">No group</option>
                  {groups.map((group) => (
                    <option key={group.id} value={group.id}>
                      {group.name}
                    </option>
                  ))}
                </select>
                <select name="program_id" defaultValue={editingStudent.program_id || ""} required>
                  <option value="">Program</option>
                  {studyPrograms.map((program) => (
                    <option key={program.id} value={program.id}>
                      {program.name}
                    </option>
                  ))}
                </select>
                <input name="full_name" defaultValue={editingStudent.full_name} required />
                <input name="phone" defaultValue={editingStudent.phone} required />
                <select name="status" defaultValue={editingStudent.status}>
                  {studentStatusOptions.map((option) => (
                    <option key={option} value={option}>
                      {option}
                    </option>
                  ))}
                </select>
                <select name="preferred_weekdays" multiple size="4" defaultValue={editingStudent.preferred_weekdays || []}>
                  {studentPreferredWeekdayOptions.map((option) => (
                    <option key={option} value={option}>
                      {option}
                    </option>
                  ))}
                </select>
                <select name="preferred_time_windows" multiple size="3" defaultValue={editingStudent.preferred_time_windows || []}>
                  {studentPreferredTimeWindowOptions.map((option) => (
                    <option key={option} value={option}>
                      {option}
                    </option>
                  ))}
                </select>
                <input
                  name="manual_hours_completed"
                  type="number"
                  min="0"
                  step="0.5"
                  defaultValue={editingStudent.manual_hours_completed ?? 0}
                  placeholder="Manual hours completed"
                />
                <button className="primary-button" type="submit">
                  Save
                </button>
              </form>
            </div>
          ) : null}

          <div className="panel">
            <div className="panel-head">
              <div>
                <span className="eyebrow">Students</span>
                <h3>Student List</h3>
              </div>
              <select value={studentGroupFilter} onChange={(event) => setStudentGroupFilter(event.target.value)}>
                <option value="">All groups</option>
                {groups.map((group) => (
                  <option key={group.id} value={group.id}>
                    {group.name}
                  </option>
                ))}
              </select>
            </div>
            <DataTable headers={["Name", "Phone", "Branch", "Group", "Program", "Manual hours", "Status", "Actions"]} emptyMessage="No students yet.">
              {filteredStudents.map((item) => (
                <tr key={item.id}>
                  <td>{item.full_name}</td>
                  <td>{item.phone}</td>
                  <td>{branchNameById[item.branch_id] || item.branch_id}</td>
                  <td>{item.group_id ? groupNameById[item.group_id] || item.group_id : "No group"}</td>
                  <td>{studyProgramNameById[item.program_id] || item.program_id || "No program"}</td>
                  <td>{item.manual_hours_completed ?? 0}</td>
                  <td><Pill>{item.status}</Pill></td>
                  <td className="table-actions">
                    <button type="button" className="ghost-button" onClick={() => setSelectedStudentId(item.id)}>
                      Focus
                    </button>
                    {canWriteWorkspace ? (
                      <button type="button" className="ghost-button" onClick={() => setEditingStudentId(item.id)}>
                        Edit
                      </button>
                    ) : null}
                    {canWriteWorkspace ? (
                      <button
                        type="button"
                        className="icon-button"
                        onClick={() => {
                          if (!window.confirm("Delete student?")) return;
                          void handleDelete(() => studentsApi(`/students/${item.id}`, { method: "DELETE" }), "Student deleted.");
                        }}
                      >
                        Delete
                      </button>
                    ) : null}
                  </td>
                </tr>
              ))}
            </DataTable>
          </div>
        </div>

        <div className="stack-column">
          {canWriteWorkspace ? (
            <div className="panel">
              <div className="panel-head">
                <div>
                  <span className="eyebrow">Programs</span>
                  <h3>Create Study Program</h3>
                </div>
              </div>
              <form className="toolbar-form" onSubmit={createStudyProgram}>
                <input name="name" placeholder="Program name" required />
                <select name="transmission" defaultValue="manual" required>
                  {transmissionTypeOptions.map((option) => (
                    <option key={option} value={option}>
                      {option}
                    </option>
                  ))}
                </select>
                <input name="total_hours" type="number" min="0" step="0.5" placeholder="Total hours" required />
                <button className="primary-button" type="submit">
                  Create
                </button>
              </form>
            </div>
          ) : null}

          {canWriteWorkspace && editingProgram ? (
            <div className="panel">
              <div className="panel-head">
                <div>
                  <span className="eyebrow">Programs</span>
                  <h3>Edit Study Program</h3>
                </div>
                <button type="button" className="ghost-button" onClick={() => setEditingProgramId("")}>
                  Hide
                </button>
              </div>
              <form key={editingProgram.id} className="toolbar-form" onSubmit={updateStudyProgram}>
                <input name="name" defaultValue={editingProgram.name || ""} placeholder="Program name" required />
                <select name="transmission" defaultValue={editingProgram.transmission || "manual"} required>
                  {transmissionTypeOptions.map((option) => (
                    <option key={option} value={option}>
                      {option}
                    </option>
                  ))}
                </select>
                <input
                  name="total_hours"
                  type="number"
                  min="0"
                  step="0.5"
                  defaultValue={editingProgram.total_hours ?? 0}
                  placeholder="Total hours"
                  required
                />
                <button className="primary-button" type="submit">
                  Save
                </button>
              </form>
            </div>
          ) : null}

          <div className="panel">
            <div className="panel-head">
              <div>
                <span className="eyebrow">Programs</span>
                <h3>Study Program List</h3>
              </div>
            </div>
            <DataTable headers={["Name", "Transmission", "Total hours", "Created", "Actions"]} emptyMessage="No study programs yet.">
              {studyPrograms.map((item) => (
                <tr key={item.id}>
                  <td>{item.name}</td>
                  <td>{item.transmission}</td>
                  <td>{item.total_hours}</td>
                  <td>{formatDateTime(item.created_at)}</td>
                  <td className="table-actions">
                    {canWriteWorkspace ? (
                      <button type="button" className="ghost-button" onClick={() => setEditingProgramId(item.id)}>
                        Edit
                      </button>
                    ) : null}
                    {canWriteWorkspace ? (
                      <button
                        type="button"
                        className="icon-button"
                        onClick={() => {
                          if (!window.confirm("Delete study program?")) return;
                          void handleDelete(() => studentsApi(`/study-programs/${item.id}`, { method: "DELETE" }), "Study program deleted.");
                        }}
                      >
                        Delete
                      </button>
                    ) : null}
                  </td>
                </tr>
              ))}
            </DataTable>
          </div>

          {canWriteWorkspace ? (
            <div className="panel">
              <div className="panel-head">
                <div>
                  <span className="eyebrow">Groups</span>
                  <h3>Create Group</h3>
                </div>
              </div>
              <form className="toolbar-form" onSubmit={createGroup}>
                <BranchSelect branches={branches} name="branch_id" includePlaceholder />
                <input name="name" placeholder="Group name" required />
                <button className="primary-button" type="submit">
                  Create
                </button>
              </form>
            </div>
          ) : null}

          {canWriteWorkspace && editingGroup ? (
            <div className="panel">
              <div className="panel-head">
                <div>
                  <span className="eyebrow">Groups</span>
                  <h3>Edit Group</h3>
                </div>
                <button type="button" className="ghost-button" onClick={() => setEditingGroupId("")}>
                  Hide
                </button>
              </div>
              <form key={editingGroup.id} className="toolbar-form" onSubmit={updateGroup}>
                <select name="branch_id" defaultValue={editingGroup.branch_id}>
                  {branches.map((branch) => (
                    <option key={branch.id} value={branch.id}>
                      {branch.name}
                    </option>
                  ))}
                </select>
                <input name="name" defaultValue={editingGroup.name} required />
                <button className="primary-button" type="submit">
                  Save
                </button>
              </form>
            </div>
          ) : null}

          <div className="panel">
            <div className="panel-head">
              <div>
                <span className="eyebrow">Groups</span>
                <h3>Group List</h3>
              </div>
            </div>
            <DataTable headers={["Name", "Branch", "Created", "Actions"]} emptyMessage="No groups yet.">
              {groups.map((item) => (
                <tr key={item.id}>
                  <td>{item.name}</td>
                  <td>{branchNameById[item.branch_id] || item.branch_id}</td>
                  <td>{formatDateTime(item.created_at)}</td>
                  <td className="table-actions">
                    {canWriteWorkspace ? (
                      <button type="button" className="ghost-button" onClick={() => setEditingGroupId(item.id)}>
                        Edit
                      </button>
                    ) : null}
                    {canWriteWorkspace ? (
                      <button
                        type="button"
                        className="icon-button"
                        onClick={() => {
                          if (!window.confirm("Delete group?")) return;
                          void handleDelete(() => studentsApi(`/groups/${item.id}`, { method: "DELETE" }), "Group deleted.");
                        }}
                      >
                        Delete
                      </button>
                    ) : null}
                  </td>
                </tr>
              ))}
            </DataTable>
          </div>
        </div>
      </section>
    );
  }

  function renderLessonsTabV2() {
    return (
      <section className="stack-column">
        {canWriteWorkspace ? (
          <div className="panel">
            <div className="panel-head">
              <div>
                <span className="eyebrow">Lessons</span>
                <h3>Create Lesson</h3>
              </div>
            </div>
            <form className="toolbar-form wide" onSubmit={createLesson}>
              <BranchSelect branches={branches} name="branch_id" includePlaceholder />
              <select name="student_id" defaultValue="">
                <option value="">Student</option>
                {students.map((item) => (
                  <option key={item.id} value={item.id}>
                    {item.full_name} - {groupNameById[item.group_id] || "No group"}
                  </option>
                ))}
              </select>
              <select name="instructor_id" defaultValue="">
                <option value="">Instructor</option>
                {instructors.map((item) => (
                  <option key={item.id} value={item.id}>
                    {item.full_name}
                  </option>
                ))}
              </select>
              <select name="car_id" defaultValue="">
                <option value="">Car</option>
                {cars.map((item) => (
                  <option key={item.id} value={item.id}>
                    {carNameById[item.id]}
                  </option>
                ))}
              </select>
              <input name="start_at" type="datetime-local" defaultValue={nextDateTime(1)} />
              <input name="end_at" type="datetime-local" defaultValue={nextDateTime(3)} />
              <select name="status" defaultValue="scheduled">
                {lessonStatusOptions.map((option) => (
                  <option key={option} value={option}>
                    {option}
                  </option>
                ))}
              </select>
              <input name="comment" placeholder="Comment" />
              <button className="primary-button" type="submit">
                Create
              </button>
            </form>
          </div>
        ) : null}

        {canWriteWorkspace && editingLesson ? (
          <div className="panel">
            <div className="panel-head">
              <div>
                <span className="eyebrow">Lessons</span>
                <h3>Edit Lesson</h3>
              </div>
              <button type="button" className="ghost-button" onClick={() => setEditingLessonId("")}>
                Hide
              </button>
            </div>
            <form key={editingLesson.id} className="toolbar-form wide" onSubmit={updateLesson}>
              <select name="branch_id" defaultValue={editingLesson.branch_id || ""}>
                <option value="">No branch</option>
                {branches.map((branch) => (
                  <option key={branch.id} value={branch.id}>
                    {branch.name}
                  </option>
                ))}
              </select>
              <select name="student_id" defaultValue={editingLesson.student_id}>
                {students.map((item) => (
                  <option key={item.id} value={item.id}>
                    {item.full_name} - {groupNameById[item.group_id] || "No group"}
                  </option>
                ))}
              </select>
              <select name="instructor_id" defaultValue={editingLesson.instructor_id}>
                {instructors.map((item) => (
                  <option key={item.id} value={item.id}>
                    {item.full_name}
                  </option>
                ))}
              </select>
              <select name="car_id" defaultValue={editingLesson.car_id}>
                {cars.map((item) => (
                  <option key={item.id} value={item.id}>
                    {carNameById[item.id]}
                  </option>
                ))}
              </select>
              <input name="start_at" type="datetime-local" defaultValue={asLocalDateTime(editingLesson.start_at)} />
              <input name="end_at" type="datetime-local" defaultValue={asLocalDateTime(editingLesson.end_at)} />
              <select name="status" defaultValue={editingLesson.status}>
                {lessonStatusOptions.map((option) => (
                  <option key={option} value={option}>
                    {option}
                  </option>
                ))}
              </select>
              <input name="comment" defaultValue={editingLesson.comment || ""} placeholder="Comment" />
              <button className="primary-button" type="submit">
                Save
              </button>
            </form>
          </div>
        ) : null}

        <div className="panel">
          <div className="panel-head">
            <div>
              <span className="eyebrow">Lessons</span>
              <h3>Lesson List</h3>
            </div>
          </div>
          <DataTable headers={["Student", "Group", "Instructor", "Car", "Start", "End", "Status", "Actions"]} emptyMessage="No lessons yet.">
            {lessons.map((item) => (
              <tr key={item.id}>
                <td>{studentNameById[item.student_id] || item.student_id}</td>
                <td>{groupNameById[students.find((student) => student.id === item.student_id)?.group_id] || "No group"}</td>
                <td>{instructorNameById[item.instructor_id] || item.instructor_id}</td>
                <td>{carNameById[item.car_id] || item.car_id}</td>
                <td>{formatDateTime(item.start_at)}</td>
                <td>{formatDateTime(item.end_at)}</td>
                <td><Pill>{item.status}</Pill></td>
                <td className="table-actions">
                  {canWriteWorkspace ? (
                    <button type="button" className="ghost-button" onClick={() => setEditingLessonId(item.id)}>
                      Edit
                    </button>
                  ) : null}
                  {canWriteWorkspace ? (
                    <button
                      type="button"
                      className="icon-button"
                      onClick={() => {
                        if (!window.confirm("Delete lesson?")) return;
                        void handleDelete(() => lessonsApi(`/lessons/${item.id}`, { method: "DELETE" }), "Lesson deleted.");
                      }}
                    >
                      Delete
                    </button>
                  ) : null}
                </td>
              </tr>
            ))}
          </DataTable>
        </div>
      </section>
    );
  }

  function renderPaymentsTabV2() {
    return (
      <section className="workspace-grid two-col">
        <div className="stack-column">
          <div className="panel">
            <div className="panel-head">
              <div>
                <span className="eyebrow">Student Focus</span>
                <h3>Student Balance</h3>
              </div>
              <select value={selectedStudentId} onChange={(event) => setSelectedStudentId(event.target.value)}>
                <option value="">Select student</option>
                {students.map((item) => (
                  <option key={item.id} value={item.id}>
                    {item.full_name}
                  </option>
                ))}
              </select>
            </div>
            {studentBalance ? (
              <div className="detail-grid">
                <Detail label="Contract" value={formatMoney(studentBalance.contract_total)} />
                <Detail label="Paid" value={formatMoney(studentBalance.paid_amount)} />
                <Detail label="Remaining" value={formatMoney(studentBalance.remaining_amount)} />
                <Detail label="Student ID" value={studentBalance.student_id} />
              </div>
            ) : (
              <EmptyState message="Select a student to load the balance." />
            )}
          </div>

          {canWriteWorkspace ? (
            <div className="panel">
              <div className="panel-head">
                <div>
                  <span className="eyebrow">Payments</span>
                  <h3>Create Payment</h3>
                </div>
              </div>
              <form className="toolbar-form wide" onSubmit={createPayment}>
                <select name="student_id" defaultValue={selectedStudentId || ""}>
                  <option value="">Student</option>
                  {students.map((item) => (
                    <option key={item.id} value={item.id}>
                      {item.full_name}
                    </option>
                  ))}
                </select>
                <select name="type" defaultValue="cash">
                  {paymentTypeOptions.map((option) => (
                    <option key={option} value={option}>
                      {option}
                    </option>
                  ))}
                </select>
                <input name="amount" type="number" min="0.01" step="0.01" placeholder="Amount" required />
                <input name="comment" placeholder="Comment" />
                <button className="primary-button" type="submit">
                  Create
                </button>
              </form>
            </div>
          ) : null}

          {canWriteWorkspace ? (
            <div className="panel">
              <div className="panel-head">
                <div>
                  <span className="eyebrow">Contract</span>
                  <h3>Upsert Contract Total</h3>
                </div>
              </div>
              <form className="toolbar-form" onSubmit={updateContract}>
                <select name="student_id" defaultValue={selectedStudentId || ""}>
                  <option value="">Student</option>
                  {students.map((item) => (
                    <option key={item.id} value={item.id}>
                      {item.full_name}
                    </option>
                  ))}
                </select>
                <input name="contract_total" type="number" min="0" step="0.01" placeholder="Contract total" required />
                <button className="primary-button" type="submit">
                  Save
                </button>
              </form>
            </div>
          ) : null}

          {canWriteWorkspace && editingPayment ? (
            <div className="panel">
              <div className="panel-head">
                <div>
                  <span className="eyebrow">Payments</span>
                  <h3>Edit Payment</h3>
                </div>
                <button type="button" className="ghost-button" onClick={() => setEditingPaymentId("")}>
                  Hide
                </button>
              </div>
              <form key={editingPayment.id} className="toolbar-form wide" onSubmit={updatePayment}>
                <select name="type" defaultValue={editingPayment.type}>
                  {paymentTypeOptions.map((option) => (
                    <option key={option} value={option}>
                      {option}
                    </option>
                  ))}
                </select>
                <input name="amount" type="number" min="0.01" step="0.01" defaultValue={editingPayment.amount} required />
                <input name="comment" defaultValue={editingPayment.comment || ""} placeholder="Comment" />
                <button className="primary-button" type="submit">
                  Save
                </button>
              </form>
            </div>
          ) : null}
        </div>

        <div className="panel">
          <div className="panel-head">
            <div>
              <span className="eyebrow">Payments</span>
              <h3>Payment Journal</h3>
            </div>
          </div>
          <DataTable headers={["Student", "Type", "Amount", "Comment", "Created", "Actions"]} emptyMessage="No payments yet.">
            {payments.map((item) => (
              <tr key={item.id}>
                <td>{studentNameById[item.student_id] || item.student_id}</td>
                <td><Pill>{item.type}</Pill></td>
                <td>{formatMoney(item.amount)}</td>
                <td>{item.comment || "—"}</td>
                <td>{formatDateTime(item.created_at)}</td>
                <td className="table-actions">
                  {canWriteWorkspace ? (
                    <button type="button" className="ghost-button" onClick={() => setEditingPaymentId(item.id)}>
                      Edit
                    </button>
                  ) : null}
                  {canWriteWorkspace ? (
                    <button
                      type="button"
                      className="icon-button"
                      onClick={() => {
                        if (!window.confirm("Delete payment?")) return;
                        void handleDelete(() => paymentsApi(`/payments/${item.id}`, { method: "DELETE" }), "Payment deleted.");
                      }}
                    >
                      Delete
                    </button>
                  ) : null}
                </td>
              </tr>
            ))}
          </DataTable>
        </div>
      </section>
    );
  }

  function renderDocumentsTabV2() {
    return (
      <section className="stack-column">
        <div className="workspace-grid two-col">
          <div className="stack-column">
            {canWriteWorkspace ? (
              <div className="panel">
                <div className="panel-head">
                  <div>
                    <span className="eyebrow">Upload</span>
                    <h3>Upload Document</h3>
                  </div>
                </div>
                <form className="form-stack" onSubmit={uploadDocument}>
                  <select name="student_id" defaultValue={selectedStudentId || ""}>
                    <option value="">Student</option>
                    {students.map((item) => (
                      <option key={item.id} value={item.id}>
                        {item.full_name}
                      </option>
                    ))}
                  </select>
                  <input name="type" placeholder="Document type" required />
                  <input name="title" placeholder="Title" required />
                  <input name="file" type="file" required />
                  <button className="primary-button" type="submit">
                    Upload
                  </button>
                </form>
              </div>
            ) : null}

            {canWriteWorkspace ? (
              <div className="panel">
                <div className="panel-head">
                  <div>
                    <span className="eyebrow">Templates</span>
                    <h3>Create Template</h3>
                  </div>
                </div>
                <form className="form-stack" onSubmit={createTemplate}>
                  <input name="name" placeholder="Template name" required />
                  <input name="type" placeholder="Document type" required />
                  <textarea name="content" rows="8" placeholder="{{.full_name}}" required />
                  <button className="primary-button" type="submit">
                    Create
                  </button>
                </form>
              </div>
            ) : null}

            {canWriteWorkspace && editingTemplate ? (
              <div className="panel">
                <div className="panel-head">
                  <div>
                    <span className="eyebrow">Templates</span>
                    <h3>Edit Template</h3>
                  </div>
                  <button type="button" className="ghost-button" onClick={() => setEditingTemplateId("")}>
                    Hide
                  </button>
                </div>
                <form key={editingTemplate.id} className="form-stack" onSubmit={updateTemplate}>
                  <input name="name" defaultValue={editingTemplate.name} required />
                  <input name="type" defaultValue={editingTemplate.type} required />
                  <textarea name="content" rows="8" defaultValue={editingTemplate.content} required />
                  <button className="primary-button" type="submit">
                    Save
                  </button>
                </form>
              </div>
            ) : null}
          </div>

          <div className="stack-column">
            {canWriteWorkspace ? (
              <div className="panel">
                <div className="panel-head">
                  <div>
                    <span className="eyebrow">Generate</span>
                    <h3>Generate Document</h3>
                  </div>
                </div>
                <form className="form-stack" onSubmit={generateDocument}>
                  <select name="student_id" defaultValue={selectedStudentId || ""}>
                    <option value="">Student</option>
                    {students.map((item) => (
                      <option key={item.id} value={item.id}>
                        {item.full_name}
                      </option>
                    ))}
                  </select>
                  <select name="template_id" defaultValue="">
                    <option value="">Template</option>
                    {documentTemplates.map((item) => (
                      <option key={item.id} value={item.id}>
                        {item.name}
                      </option>
                    ))}
                  </select>
                  <input name="title" placeholder="Document title" required />
                  <input name="file_name" placeholder="Optional file name" />
                  <textarea name="placeholders" rows="7" placeholder={"full_name=Ivan Ivanov\ncontract_total=50000"} />
                  <button className="primary-button" type="submit">
                    Generate
                  </button>
                </form>
              </div>
            ) : null}

            <div className="panel">
              <div className="panel-head">
                <div>
                  <span className="eyebrow">Templates</span>
                  <h3>Template Registry</h3>
                </div>
              </div>
              <DataTable headers={["Name", "Type", "Created", "Actions"]} emptyMessage="No templates yet.">
                {documentTemplates.map((item) => (
                  <tr key={item.id}>
                    <td>{item.name}</td>
                    <td><Pill>{item.type}</Pill></td>
                    <td>{formatDateTime(item.created_at)}</td>
                    <td className="table-actions">
                      {canWriteWorkspace ? (
                        <button type="button" className="ghost-button" onClick={() => setEditingTemplateId(item.id)}>
                          Edit
                        </button>
                      ) : null}
                      {canWriteWorkspace ? (
                        <button
                          type="button"
                          className="icon-button"
                          onClick={() => {
                            if (!window.confirm("Delete template?")) return;
                            void handleDelete(
                              () => documentsApi(`/document-templates/${item.id}`, { method: "DELETE" }),
                              "Template deleted."
                            );
                          }}
                        >
                          Delete
                        </button>
                      ) : null}
                    </td>
                  </tr>
                ))}
              </DataTable>
            </div>
          </div>
        </div>

        <div className="panel">
          <div className="panel-head">
            <div>
              <span className="eyebrow">Documents</span>
              <h3>Document Registry</h3>
            </div>
          </div>

          {documentDetails ? (
            <div className="detail-grid document-detail-grid">
              <Detail label="ID" value={documentDetails.id} />
              <Detail label="Student ID" value={documentDetails.student_id} />
              <Detail label="Type" value={documentDetails.type} />
              <Detail label="Status" value={documentDetails.status} />
              <Detail label="Mime Type" value={documentDetails.mime_type} />
              <Detail label="Object Key" value={documentDetails.object_key} />
              <Detail label="Size" value={`${documentDetails.size} bytes`} />
              <Detail label="Created" value={formatDateTime(documentDetails.created_at)} />
            </div>
          ) : null}

          <DataTable headers={["Title", "Student", "Type", "Status", "Mime", "Size", "Actions"]} emptyMessage="No documents yet.">
            {documents.map((item) => (
              <tr key={item.id}>
                <td>{item.title}</td>
                <td>{studentNameById[item.student_id] || item.student_id}</td>
                <td><Pill>{item.type}</Pill></td>
                <td><Pill>{item.status}</Pill></td>
                <td>{item.mime_type}</td>
                <td>{item.size} bytes</td>
                <td className="table-actions">
                  <button type="button" className="ghost-button" onClick={() => void inspectDocument(item.id)}>
                    Details
                  </button>
                  <button type="button" className="ghost-button" onClick={() => void downloadDocument(item.id)}>
                    Download
                  </button>
                  {canWriteWorkspace ? (
                    <button
                      type="button"
                      className="icon-button"
                      onClick={() => {
                        if (!window.confirm("Delete document?")) return;
                        void handleDelete(() => documentsApi(`/documents/${item.id}`, { method: "DELETE" }), "Document deleted.");
                      }}
                    >
                      Delete
                    </button>
                  ) : null}
                </td>
              </tr>
            ))}
          </DataTable>
        </div>
      </section>
    );
  }

  function renderWorkspace() {
    const tabs = [
      ["overview", "Обзор"],
      ["resources", "Ресурсы"],
      ["students", "Ученики"],
      ["lessons", "Занятия"],
      ["payments", "Платежи"],
      ["documents", "Документы"],
      ["planning", "Planning"],
      ["reporting", "Reporting"],
    ];

    return (
      <>
        <section className="tabbar">
          {tabs.map(([id, label]) => (
            <button
              key={id}
              type="button"
              className={`tab-button ${workspaceTab === id ? "active" : ""}`}
              onClick={() => setWorkspaceTab(id)}
            >
              {label}
            </button>
          ))}
        </section>

        {workspaceTab === "overview" ? renderOverview() : null}
        {workspaceTab === "resources" ? renderResourcesTab() : null}
        {workspaceTab === "students" ? renderStudentsTabV2() : null}
        {workspaceTab === "lessons" ? renderLessonsTabV2() : null}
        {workspaceTab === "payments" ? renderPaymentsTabV2() : null}
        {workspaceTab === "documents" ? renderDocumentsTabV2() : null}
        {workspaceTab === "planning" ? renderPlanningTab() : null}
        {workspaceTab === "reporting" ? renderReportingTab() : null}
      </>
    );
  }

  function renderToast() {
    if (!toast) return null;
    return (
      <div className={`toast ${toast.type}`}>
        <span>{toast.message}</span>
        <button type="button" className="icon-link" onClick={() => setToast(null)}>
          Закрыть
        </button>
      </div>
    );
  }

  if (view === "login") {
    return renderLogin();
  }

  return (
    <div className="crm-shell">
      <aside className="sidebar">
        <div className="brand-block">
          <div className="brand-mark">M</div>
          <div>
            <span className="sidebar-label bright">Magistral</span>
            <h1>Frontend Control Room</h1>
          </div>
        </div>

        <div className="sidebar-card">
          <span className="sidebar-label">Режим</span>
          <strong>{isSuperAdmin ? "SaaS Panel" : "School Workspace"}</strong>
          <span className="muted-light">{roleName(me?.role)}</span>
        </div>

        <div className="sidebar-card">
          <span className="sidebar-label">Identity</span>
          <strong>{me?.user_id}</strong>
          <span className="muted-light">{me?.school_id || "global scope"}</span>
        </div>

        <div className="sidebar-card">
          <span className="sidebar-label">Контекст</span>
          <span className="muted-light">
            {isSuperAdmin
              ? "Управление школами, филиалами и bootstrap пользователей."
              : "Рабочее место для ресурсов, учеников, расписания, платежей, документов и reporting."}
          </span>
        </div>

        <button type="button" className="ghost-danger" onClick={handleLogout}>
          Выйти
        </button>
      </aside>

      <main className="main-area">
        <header className="page-head">
          <div>
            <span className="eyebrow">{isSuperAdmin ? "Super Admin" : "School Workspace"}</span>
            <h2>{isSuperAdmin ? "Управление платформой" : "Операционный кабинет школы"}</h2>
          </div>
          <div className="head-badges">
            {busy ? <span className="busy-pill">Синхронизация...</span> : null}
            <span className="state-pill accent">{roleName(me?.role)}</span>
            {!isSuperAdmin ? <span className="state-pill">{canWriteWorkspace ? "write access" : "read only"}</span> : null}
          </div>
        </header>

        {renderToast()}
        {isSuperAdmin ? renderSuperAdmin() : renderWorkspace()}
      </main>
    </div>
  );
}

function StatCard({ label, value, hint }) {
  return (
    <article className="kpi-card">
      <span className="eyebrow">{label}</span>
      <strong>{value}</strong>
      <p>{hint}</p>
    </article>
  );
}

function DataTable({ headers, emptyMessage, children }) {
  const rows = Array.isArray(children) ? children.filter(Boolean) : [];
  return (
    <div className="table-shell">
      <table className="crm-table">
        <thead>
          <tr>
            {headers.map((header) => (
              <th key={header}>{header}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.length ? (
            rows
          ) : (
            <tr>
              <td colSpan={headers.length}>
                <EmptyState compact message={emptyMessage} />
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
}

function Detail({ label, value }) {
  return (
    <div className="detail-card">
      <span>{label}</span>
      <strong>{value || "—"}</strong>
    </div>
  );
}

function Pill({ children }) {
  return <span className="state-pill">{children}</span>;
}

function EmptyState({ message, compact = false }) {
  return <div className={`empty-state ${compact ? "compact" : ""}`}>{message}</div>;
}

function BranchSelect({ branches, name, includePlaceholder = false }) {
  return (
    <select name={name} defaultValue="">
      {includePlaceholder ? <option value="">Выберите филиал</option> : null}
      {branches.map((branch) => (
        <option key={branch.id} value={branch.id}>
          {branch.name}
        </option>
      ))}
    </select>
  );
}

export default App;
