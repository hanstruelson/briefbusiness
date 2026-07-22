<template>
  <div class="dashboard-container">
    <!-- Sidebar / Navigation -->
    <aside class="sidebar">
      <div class="logo-area">
        <svg class="scale-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M12 3v17M12 3L4 9M12 3l8 6M4 9h16M4 9l4 7m8-7l-4 7M8 16a3 3 0 006 0" />
        </svg>
        <span class="logo-text">AppealsBrief</span>
      </div>
      <nav class="sidebar-nav">
        <a href="#" class="nav-item active">
          <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M4 6h16M4 12h16M4 18h7" />
          </svg>
          All Briefs
        </a>
      </nav>
      <div class="sidebar-footer">
        <div class="user-info">
          <div class="user-avatar">{{ userEmail[0]?.toUpperCase() || 'U' }}</div>
          <div class="user-details">
            <p class="user-name">Drafting Counsel</p>
            <p class="user-email">{{ userEmail }}</p>
          </div>
        </div>
        <button @click="handleLogout" class="btn-logout">
          Sign Out
        </button>
      </div>
    </aside>

    <!-- Main Content Area -->
    <main class="main-content">
      <header class="content-header">
        <div>
          <h1>Appellate Briefs</h1>
          <p class="subtitle">Draft, search, and generate high-quality court appeals</p>
        </div>
        <button @click="createNewDraft" class="btn-primary">
          <svg class="btn-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 5v14M5 12h14" />
          </svg>
          Create New Brief
        </button>
      </header>

      <!-- Search Section -->
      <section class="search-section">
        <div class="search-bar-wrapper">
          <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8" />
            <path d="M21 21l-4.35-4.35" />
          </svg>
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search briefs by title, jurisdiction, case number, or content..."
            class="search-input"
          />
        </div>
      </section>

      <!-- Drafts Grid -->
      <div v-if="isLoading" class="loader-container">
        <div class="spinner"></div>
        <p>Loading your brief portfolio...</p>
      </div>

      <div v-else-if="filteredDrafts.length === 0" class="empty-state">
        <div class="empty-icon-wrapper">
          <svg class="empty-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z" />
            <path d="M14 2v6h6M16 13H8M16 17H8M10 9H8" />
          </svg>
        </div>
        <h3>No briefs found</h3>
        <p v-if="searchQuery">No results match your search query. Try typing something else.</p>
        <p v-else>Get started by creating your first appellate brief.</p>
        <button v-if="!searchQuery" @click="createNewDraft" class="btn-secondary">
          Create New Brief
        </button>
      </div>

      <div v-else class="drafts-grid">
        <article v-for="item in filteredDrafts" :key="item.draft.id" class="draft-card">
          <div class="card-header">
            <div class="case-meta-tag" v-if="parseMeta(item.draft.markdown).caseNumber">
              No. {{ parseMeta(item.draft.markdown).caseNumber }}
            </div>
            <span class="card-date">{{ formatDate(item.draft.updated_at) }}</span>
          </div>
          
          <h2 class="card-title">{{ item.draft.title }}</h2>
          
          <p class="card-jurisdiction" v-if="parseMeta(item.draft.markdown).jurisdiction">
            {{ parseMeta(item.draft.markdown).jurisdiction }}
          </p>

          <p class="card-snippet">
            {{ getSnippet(item.draft.markdown) }}
          </p>

          <div class="card-footer">
            <div class="parties-names" v-if="parseMeta(item.draft.markdown).appellant">
              <span>{{ parseMeta(item.draft.markdown).appellant.split(',')[0] }}</span>
              <span class="vs">v.</span>
              <span>{{ parseMeta(item.draft.markdown).appellee.split(',')[0] }}</span>
            </div>
            <div class="card-actions">
              <button @click="downloadPDF(item.draft.id)" class="btn-icon-only" title="Download PDF">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M7 10l5 5 5-5M12 15V3" />
                </svg>
              </button>
              <button @click="openDraft(item.draft.id)" class="btn-card-primary">
                Open Editor
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M5 12h14M12 5l7 7-7 7" />
                </svg>
              </button>
            </div>
          </div>
        </article>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import { searchDrafts } from '../utilities/search';

const router = useRouter();
const drafts = ref([]);
const searchQuery = ref('');
const isLoading = ref(true);
const userEmail = ref('counsel@court.gov');

// Check authentication
onMounted(async () => {
  const token = localStorage.getItem('AccessToken');
  if (!token) {
    router.push('/register');
    return;
  }
  
  // Try retrieving user details or fall back
  try {
    // Basic load drafts
    await loadDrafts();
  } catch (error) {
    console.error('Failed to load dashboard:', error);
  } finally {
    isLoading.value = false;
  }
});

const loadDrafts = async () => {
  const token = localStorage.getItem('AccessToken');
  const response = await fetch('/api/drafts', {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  if (response.ok) {
    drafts.value = await response.json();
  } else if (response.status === 401) {
    localStorage.removeItem('AccessToken');
    router.push('/register');
  }
};

const filteredDrafts = computed(() => {
  return searchDrafts(drafts.value, searchQuery.value);
});

const parseMeta = (markdown) => {
  if (!markdown) return { jurisdiction: '', caseNumber: '', appellant: '', appellee: '' };
  
  // Extract JSON frontmatter between ---json and ---
  if (markdown.startsWith('---json')) {
    const endIdx = markdown.indexOf('---', 7);
    if (endIdx !== -1) {
      try {
        const rawJson = markdown.substring(7, endIdx).trim();
        return JSON.parse(rawJson);
      } catch (e) {
        // ignore
      }
    }
  } else if (markdown.startsWith('---')) {
    const endIdx = markdown.indexOf('---', 3);
    if (endIdx !== -1) {
      const rawFM = markdown.substring(3, endIdx).trim();
      const meta = { jurisdiction: '', caseNumber: '', appellant: '', appellee: '' };
      const lines = rawFM.split('\n');
      for (const line of lines) {
        const colonIdx = line.indexOf(':');
        if (colonIdx !== -1) {
          const key = line.substring(0, colonIdx).trim();
          const val = line.substring(colonIdx + 1).trim();
          if (key in meta) {
            meta[key] = val.replace(/^["']|["']$/g, ''); // strip quotes
          }
        }
      }
      return meta;
    }
  }
  return { jurisdiction: '', caseNumber: '', appellant: '', appellee: '' };
};

const getSnippet = (markdown) => {
  if (!markdown) return '';
  // Strip frontmatter
  let cleanText = markdown;
  if (markdown.startsWith('---')) {
    const endIdx = markdown.indexOf('---', 3);
    if (endIdx !== -1) {
      cleanText = markdown.substring(endIdx + 3).trim();
    }
  }
  
  // Remove markdown headers
  cleanText = cleanText.replace(/#+\s+.*?\n/g, ' ');
  // Remove markdown brackets
  cleanText = cleanText.replace(/\[\[(.*?)\]\]/g, '$1');
  
  // Truncate to 140 chars
  if (cleanText.length > 140) {
    return cleanText.substring(0, 137).trim() + '...';
  }
  return cleanText;
};

const formatDate = (dateStr) => {
  if (!dateStr) return '';
  const date = new Date(dateStr);
  return date.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric'
  }) + ' at ' + date.toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit'
  });
};

const openDraft = (id) => {
  router.push(`/drafts/${id}`);
};

const createNewDraft = () => {
  router.push('/drafts/new');
};

const downloadPDF = async (id) => {
  const token = localStorage.getItem('AccessToken');
  window.open(`/api/drafts/${id}/pdf?token=${token}`, '_blank');
};

const handleLogout = () => {
  localStorage.removeItem('AccessToken');
  localStorage.removeItem('UserId');
  router.push('/register');
};
</script>

<style scoped>
.dashboard-container {
  display: grid;
  grid-template-columns: 260px 1fr;
  min-height: 100vh;
  background-color: #f8fafc;
  color: #0f172a;
}

/* Sidebar */
.sidebar {
  background-color: #0f172a;
  color: #f1f5f9;
  display: flex;
  flex-direction: column;
  padding: 24px;
  border-right: 1px solid #1e293b;
}

.logo-area {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 36px;
}

.scale-icon {
  width: 28px;
  height: 28px;
  color: #fbbf24; /* legal gold */
}

.logo-text {
  font-size: 1.25rem;
  font-weight: 700;
  letter-spacing: -0.5px;
}

.sidebar-nav {
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex-grow: 1;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-radius: 8px;
  color: #94a3b8;
  text-decoration: none;
  font-weight: 500;
  transition: all 0.2s ease;
}

.nav-item:hover, .nav-item.active {
  color: #ffffff;
  background-color: #1e293b;
}

.nav-icon {
  width: 20px;
  height: 20px;
}

.sidebar-footer {
  border-top: 1px solid #1e293b;
  padding-top: 20px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.user-avatar {
  width: 38px;
  height: 38px;
  background-color: #fbbf24;
  color: #0f172a;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
}

.user-details {
  overflow: hidden;
}

.user-name {
  font-size: 0.85rem;
  font-weight: 600;
  margin: 0;
}

.user-email {
  font-size: 0.75rem;
  color: #64748b;
  margin: 0;
  text-overflow: ellipsis;
  overflow: hidden;
  white-space: nowrap;
}

.btn-logout {
  width: 100%;
  padding: 10px;
  background: transparent;
  border: 1px solid #334155;
  color: #94a3b8;
  border-radius: 8px;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-logout:hover {
  background-color: #1e293b;
  color: #ffffff;
  border-color: #475569;
}

/* Main Content */
.main-content {
  padding: 40px;
  overflow-y: auto;
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
}

.content-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 32px;
}

.content-header h1 {
  font-size: 2.25rem;
  font-weight: 800;
  letter-spacing: -1px;
  color: #0f172a;
  margin: 0 0 6px 0;
}

.subtitle {
  color: #64748b;
  margin: 0;
  font-size: 0.95rem;
}

.btn-primary {
  padding: 12px 20px;
  background: linear-gradient(135deg, #1e3a8a 0%, #0f172a 100%);
  color: #ffffff;
  border: none;
  border-radius: 10px;
  font-size: 0.95rem;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  box-shadow: 0 4px 14px rgba(30, 58, 138, 0.25);
  transition: all 0.2s ease;
}

.btn-primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 20px rgba(30, 58, 138, 0.35);
}

.btn-icon {
  width: 18px;
  height: 18px;
}

/* Search */
.search-section {
  margin-bottom: 32px;
}

.search-bar-wrapper {
  position: relative;
  width: 100%;
}

.search-icon {
  position: absolute;
  left: 18px;
  top: 50%;
  transform: translateY(-50%);
  width: 20px;
  height: 20px;
  color: #94a3b8;
}

.search-input {
  width: 100%;
  padding: 16px 16px 16px 54px;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  font-size: 1rem;
  background-color: #ffffff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  transition: all 0.2s ease;
}

.search-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.1), 0 4px 12px rgba(0, 0, 0, 0.05);
}

/* Loader */
.loader-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 0;
  color: #64748b;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid #e2e8f0;
  border-top-color: #fbbf24;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-bottom: 16px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Empty State */
.empty-state {
  text-align: center;
  padding: 60px 40px;
  background-color: #ffffff;
  border-radius: 16px;
  border: 2px dashed #e2e8f0;
  margin-top: 20px;
}

.empty-icon-wrapper {
  width: 64px;
  height: 64px;
  background-color: #f1f5f9;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 20px;
}

.empty-icon {
  width: 32px;
  height: 32px;
  color: #94a3b8;
}

.empty-state h3 {
  font-size: 1.25rem;
  font-weight: 700;
  margin-bottom: 8px;
}

.empty-state p {
  color: #64748b;
  max-width: 400px;
  margin: 0 auto 24px;
}

.btn-secondary {
  padding: 10px 20px;
  background-color: #f1f5f9;
  color: #334155;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-secondary:hover {
  background-color: #e2e8f0;
}

/* Drafts Grid */
.drafts-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 24px;
}

.draft-card {
  background-color: #ffffff;
  border: 1px solid #edf2f7;
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05), 0 2px 4px -1px rgba(0, 0, 0, 0.03);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  display: flex;
  flex-direction: column;
}

.draft-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.08), 0 4px 6px -2px rgba(0, 0, 0, 0.04);
  border-color: #cbd5e1;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.case-meta-tag {
  background-color: #fef3c7; /* light amber */
  color: #92400e;
  padding: 4px 8px;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.5px;
}

.card-date {
  font-size: 0.8rem;
  color: #94a3b8;
}

.card-title {
  font-size: 1.35rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0 0 6px 0;
  line-height: 1.3;
}

.card-jurisdiction {
  font-size: 0.8rem;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  font-weight: 600;
  margin: 0 0 14px 0;
}

.card-snippet {
  color: #475569;
  font-size: 0.95rem;
  line-height: 1.6;
  margin: 0 0 20px 0;
  flex-grow: 1;
}

.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-top: 1px solid #f1f5f9;
  padding-top: 16px;
  margin-top: auto;
  gap: 16px;
}

.parties-names {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.85rem;
  font-weight: 600;
  color: #334155;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.vs {
  font-size: 0.75rem;
  color: #94a3b8;
  font-style: italic;
  font-weight: 400;
}

.card-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-left: auto;
  flex-shrink: 0;
}

.btn-icon-only {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
  background-color: #ffffff;
  color: #64748b;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-icon-only:hover {
  background-color: #f8fafc;
  color: #0f172a;
  border-color: #cbd5e1;
}

.btn-icon-only svg {
  width: 20px;
  height: 20px;
}

.btn-card-primary {
  padding: 10px 16px;
  background-color: #f1f5f9;
  color: #1e3a8a;
  border: none;
  border-radius: 10px;
  font-weight: 600;
  font-size: 0.9rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all 0.2s ease;
}

.btn-card-primary:hover {
  background-color: #e2e8f0;
  color: #0f172a;
}

.btn-card-primary svg {
  width: 16px;
  height: 16px;
}

@media (max-width: 768px) {
  .dashboard-container {
    grid-template-columns: 1fr;
  }
  .sidebar {
    display: none;
  }
  .main-content {
    padding: 20px;
  }
  .card-footer {
    flex-direction: column;
    align-items: flex-start;
  }
  .card-actions {
    width: 100%;
    justify-content: flex-end;
  }
}
</style>
