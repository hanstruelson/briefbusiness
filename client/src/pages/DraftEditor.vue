<template>
  <div class="editor-layout">
    <!-- Sticky Left Table of Contents -->
    <aside class="toc-sidebar">
      <div class="back-link-wrapper">
        <button @click="goBack" class="btn-back">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M19 12H5M12 19l-7-7 7-7" />
          </svg>
          Back to Dashboard
        </button>
      </div>

      <div class="toc-header">
        <h3>BRIEF SECTIONS</h3>
      </div>
      
      <nav class="toc-menu">
        <a 
          v-for="sec in tocSections" 
          :key="sec.id" 
          :href="'#' + sec.id" 
          @click.prevent="scrollToSection(sec.id)"
          class="toc-item"
          :class="{ active: activeSection === sec.id }"
        >
          <span class="toc-bullet"></span>
          <span class="toc-text">{{ sec.title }}</span>
        </a>
      </nav>

      <!-- Live Citations Display inside TOC -->
      <div class="citations-panel">
        <div class="citations-header">
          <h4>INDEX OF AUTHORITIES (c)</h4>
          <span class="citations-count">{{ liveCitations.length }}</span>
        </div>
        <div v-if="liveCitations.length === 0" class="citations-empty">
          No case citations found. Use [[Case Name, Citation]] in the body to index them.
        </div>
        <ul v-else class="citations-list">
          <li v-for="(cit, idx) in liveCitations" :key="idx" class="citation-item">
            <i>{{ splitCitation(cit).caseName }}</i>{{ splitCitation(cit).remainder }}
          </li>
        </ul>
      </div>
    </aside>

    <!-- Main Content & Scroll Container -->
    <main class="editor-main">
      <header class="editor-header">
        <div class="title-input-wrapper">
          <input 
            v-model="briefTitle" 
            type="text" 
            placeholder="Enter Appeal Brief Title (e.g. Appellant's Opening Brief)" 
            class="brief-title-input"
          />
        </div>
        <div class="header-actions">
          <button @click="triggerImport" class="btn-action-secondary" title="Import Markdown">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M12 3v12M8 7l4-4 4 4" />
            </svg>
            Import MD
          </button>
          <input 
            ref="importFileInput" 
            type="file" 
            accept=".md" 
            class="hidden-file-input" 
            @change="handleImportFile" 
          />
          <button @click="exportMarkdown" class="btn-action-secondary" title="Export Markdown">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M7 10l5 5 5-5M12 15V3" />
            </svg>
            Export MD
          </button>
          <button @click="downloadPDF" class="btn-action-gold" :disabled="!draftId">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z" />
              <path d="M14 2v6h6M16 13H8M16 17H8M10 9H8" />
            </svg>
            Download PDF
          </button>
          <button @click="saveBrief" class="btn-action-primary" :disabled="isSaving">
            {{ isSaving ? 'Saving...' : 'Save Draft' }}
          </button>
        </div>
      </header>

      <div class="editor-content-scroller" ref="scrollerRef" @scroll="onMainScroll">
        <!-- Notification Banner -->
        <div v-if="successMsg" class="success-banner">
          {{ successMsg }}
        </div>
        <div v-if="errorMsg" class="error-banner">
          {{ errorMsg }}
        </div>

        <!-- Cover Page Section -->
        <section id="cover" class="editor-section-card">
          <div class="card-section-title">
            <h2>Cover Page Details</h2>
          </div>
          
          <div class="form-grid">
            <div class="form-group full-width">
              <label>Court Jurisdiction</label>
              <input 
                v-model="metadata.jurisdiction" 
                type="text" 
                placeholder="e.g. IN THE COURT OF APPEALS FOR THE STATE OF CALIFORNIA" 
                class="form-control"
              />
            </div>
            <div class="form-group full-width">
              <label>Case Number</label>
              <input 
                v-model="metadata.caseNumber" 
                type="text" 
                placeholder="e.g. No. 2026-AP-0042" 
                class="form-control"
              />
            </div>
            <div class="form-group">
              <label>Appellant Name(s)</label>
              <textarea 
                v-model="metadata.appellant" 
                rows="3" 
                placeholder="e.g. JOHN DOE, Appellant" 
                class="form-control"
              ></textarea>
            </div>
            <div class="form-group">
              <label>Appellee Name(s)</label>
              <textarea 
                v-model="metadata.appellee" 
                rows="3" 
                placeholder="e.g. STATE OF CALIFORNIA, Appellee" 
                class="form-control"
              ></textarea>
            </div>
          </div>

          <!-- Counsel List Manager -->
          <div class="counsel-manager">
            <div class="counsel-manager-header">
              <h3>Appearing Counsel</h3>
              <button @click="addCounsel" class="btn-small-primary">
                + Add Counsel
              </button>
            </div>

            <div v-if="metadata.counselList.length === 0" class="counsel-empty">
              No counsel added yet. Counsel will appear on the cover of the PDF.
            </div>

            <div v-else class="counsel-grid-list">
              <div v-for="(c, idx) in metadata.counselList" :key="idx" class="counsel-card">
                <div class="counsel-card-header">
                  <h4>Counsel #{{ idx + 1 }}</h4>
                  <button @click="removeCounsel(idx)" class="btn-remove-counsel" title="Remove counsel">
                    Remove
                  </button>
                </div>
                <div class="form-grid-mini">
                  <div class="form-group-mini">
                    <label>Full Name</label>
                    <input v-model="c.name" type="text" class="form-control-mini" />
                  </div>
                  <div class="form-group-mini">
                    <label>Firm / Office</label>
                    <input v-model="c.firm" type="text" class="form-control-mini" />
                  </div>
                  <div class="form-group-mini full-width">
                    <label>Mailing Address</label>
                    <input v-model="c.address" type="text" class="form-control-mini" />
                  </div>
                  <div class="form-group-mini">
                    <label>Telephone Number</label>
                    <input v-model="c.phone" type="text" class="form-control-mini" />
                  </div>
                  <div class="form-group-mini">
                    <label>Email Address</label>
                    <input v-model="c.email" type="text" class="form-control-mini" />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- Section (a) -->
        <section id="sec_a" class="editor-section-card">
          <div class="card-section-title">
            <h2>(a) Identity of Parties and Counsel</h2>
            <span class="info-tip">Must list all parties to the trial court judgment and all appearing counsel details.</span>
          </div>
          <textarea 
            v-model="sections.a" 
            placeholder="Type Section (a) content..." 
            class="section-textarea"
          ></textarea>
        </section>

        <!-- Section (b) Preview -->
        <section id="sec_b" class="editor-section-card read-only-section">
          <div class="card-section-title">
            <h2>(b) Table of Contents</h2>
            <span class="info-tip">Automatically compiled. Preview of generated layout below:</span>
          </div>
          <div class="toc-preview-box">
            <div class="toc-preview-row"><span>(a) Identity of Parties and Counsel</span><span class="dot-leaders"></span><span>Page iv</span></div>
            <div class="toc-preview-row"><span>(b) Table of Contents</span><span class="dot-leaders"></span><span>Page ii</span></div>
            <div class="toc-preview-row"><span>(c) Index of Authorities</span><span class="dot-leaders"></span><span>Page iii</span></div>
            <div class="toc-preview-row"><span>(d) Statement of the Case</span><span class="dot-leaders"></span><span>Page 1</span></div>
            <div class="toc-preview-row"><span>(e) Any Statement Regarding Oral Argument</span><span class="dot-leaders"></span><span>Page 2</span></div>
            <div class="toc-preview-row"><span>(f) Issues Presented</span><span class="dot-leaders"></span><span>Page 3</span></div>
            <div class="toc-preview-row"><span>(g) Statement of Facts</span><span class="dot-leaders"></span><span>Page 4</span></div>
            <div class="toc-preview-row"><span>(h) Summary of the Argument</span><span class="dot-leaders"></span><span>Page 6</span></div>
            <div class="toc-preview-row"><span>(i) Argument</span><span class="dot-leaders"></span><span>Page 7</span></div>
            <div class="toc-preview-row"><span>(j) Prayer</span><span class="dot-leaders"></span><span>Page 12</span></div>
            <div class="toc-preview-row"><span>(k) Appendix in Civil Cases</span><span class="dot-leaders"></span><span>Page 13</span></div>
          </div>
        </section>

        <!-- Section (c) Preview -->
        <section id="sec_c" class="editor-section-card read-only-section">
          <div class="card-section-title">
            <h2>(c) Index of Authorities</h2>
            <span class="info-tip">Automatically compiled from your case citations. Preview below:</span>
          </div>
          <div class="ioa-preview-box">
            <div v-if="liveCitations.length === 0" class="no-cit-preview">
              No citations found. Write citations in double brackets, e.g. <code>[[Miranda v. Arizona, 384 U.S. 436 (1966)]]</code>, anywhere in your brief.
            </div>
            <div v-else v-for="(cit, idx) in liveCitations" :key="idx" class="toc-preview-row">
              <span>
                <i>{{ splitCitation(cit).caseName }}</i>{{ splitCitation(cit).remainder }}
              </span>
              <span class="dot-leaders"></span>
              <span>1</span>
            </div>
          </div>
        </section>

        <!-- Section (d) -->
        <section id="sec_d" class="editor-section-card">
          <div class="card-section-title">
            <h2>(d) Statement of the Case</h2>
            <span class="info-tip">State concisely the nature of thecase, course of proceedings, and trial court's disposition. Limit to 1/2 page. Do not discuss facts.</span>
          </div>
          <textarea v-model="sections.d" placeholder="Type Section (d) content..." class="section-textarea"></textarea>
        </section>

        <!-- Section (e) -->
        <section id="sec_e" class="editor-section-card">
          <div class="card-section-title">
            <h2>(e) Statement Regarding Oral Argument</h2>
            <span class="info-tip">Explain why oral argument should or should not be permitted. Maximum 1 page.</span>
          </div>
          <textarea v-model="sections.e" placeholder="Type Section (e) content..." class="section-textarea"></textarea>
        </section>

        <!-- Section (f) -->
        <section id="sec_f" class="editor-section-card">
          <div class="card-section-title">
            <h2>(f) Issues Presented</h2>
            <span class="info-tip">Concise statement of all issues or points presented for review.</span>
          </div>
          <textarea v-model="sections.f" placeholder="Type Section (f) content..." class="section-textarea"></textarea>
        </section>

        <!-- Section (g) -->
        <section id="sec_g" class="editor-section-card">
          <div class="card-section-title">
            <h2>(g) Statement of Facts</h2>
            <span class="info-tip">State concisely and without argument the facts pertinent to the issues. Add record references where possible.</span>
          </div>
          <textarea v-model="sections.g" placeholder="Type Section (g) content..." class="section-textarea"></textarea>
        </section>

        <!-- Section (h) -->
        <section id="sec_h" class="editor-section-card">
          <div class="card-section-title">
            <h2>(h) Summary of the Argument</h2>
            <span class="info-tip">A succinct, clear, and accurate summary of the arguments. Do not merely repeat the issues.</span>
          </div>
          <textarea v-model="sections.h" placeholder="Type Section (h) content..." class="section-textarea"></textarea>
        </section>

        <!-- Section (i) -->
        <section id="sec_i" class="editor-section-card">
          <div class="card-section-title">
            <h2>(i) Argument</h2>
            <span class="info-tip">The main body of the brief. State clear and concise contentions with appropriate citations.</span>
          </div>
          <textarea v-model="sections.i" placeholder="Type Section (i) content..." class="section-textarea"></textarea>
        </section>

        <!-- Section (j) -->
        <section id="sec_j" class="editor-section-card">
          <div class="card-section-title">
            <h2>(j) Prayer</h2>
            <span class="info-tip">Short conclusion stating clearly the nature of the relief sought.</span>
          </div>
          <textarea v-model="sections.j" placeholder="Type Section (j) content..." class="section-textarea"></textarea>
        </section>

        <!-- Section (k) -->
        <section id="sec_k" class="editor-section-card">
          <div class="card-section-title">
            <h2>(k) Appendix in Civil Cases</h2>
            <span class="info-tip">Must contain judgment/order appealed, jury charge/verdict, findings of fact/conclusions of law, etc.</span>
          </div>
          <textarea v-model="sections.k" placeholder="Type Section (k) content..." class="section-textarea"></textarea>
        </section>

        <!-- Version History Section -->
        <section id="history" class="editor-section-card history-section" v-if="draftId && versionHistory.length > 0">
          <div class="card-section-title">
            <h2>Revision Backups</h2>
            <span class="info-tip">Each save copies the previous version to a backup automatically. Click a version to restore it.</span>
          </div>
          <div class="history-list-box">
            <div v-for="b in versionHistory" :key="b.id" class="history-row">
              <div class="history-info">
                <span class="history-date">{{ formatDateTime(b.saved_at) }}</span>
                <span class="history-title">{{ b.title }}</span>
              </div>
              <button @click="restoreBackup(b.id)" class="btn-restore">
                Restore Version
              </button>
            </div>
          </div>
        </section>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

const route = useRoute();
const router = useRouter();
const scrollerRef = ref(null);
const importFileInput = ref(null);

const draftId = ref(0);
const briefTitle = ref('Untitled Appellate Brief');
const isSaving = ref(false);
const successMsg = ref('');
const errorMsg = ref('');
const versionHistory = ref([]);
const activeSection = ref('cover');

const metadata = reactive({
  jurisdiction: '',
  caseNumber: '',
  appellant: '',
  appellee: '',
  counselList: []
});

const sections = reactive({
  a: '',
  d: '',
  e: '',
  f: '',
  g: '',
  h: '',
  i: '',
  j: '',
  k: ''
});

const tocSections = [
  { id: 'cover', title: 'Cover Page' },
  { id: 'sec_a', title: '(a) Identity of Parties & Counsel' },
  { id: 'sec_b', title: '(b) Table of Contents (Auto)' },
  { id: 'sec_c', title: '(c) Index of Authorities (Auto)' },
  { id: 'sec_d', title: '(d) Statement of the Case' },
  { id: 'sec_e', title: '(e) Oral Argument Statement' },
  { id: 'sec_f', title: '(f) Issues Presented' },
  { id: 'sec_g', title: '(g) Statement of Facts' },
  { id: 'sec_h', title: '(h) Summary of Argument' },
  { id: 'sec_i', title: '(i) Argument' },
  { id: 'sec_j', title: '(j) Prayer' },
  { id: 'sec_k', title: '(k) Appendix' },
  { id: 'history', title: 'Revision Backups' }
];

// Load Draft Details
onMounted(async () => {
  const token = localStorage.getItem('AccessToken');
  if (!token) {
    router.push('/register');
    return;
  }

  const idParam = route.params.id;
  if (idParam && idParam !== 'new') {
    draftId.value = parseInt(idParam);
    await loadDraft();
  }
});

const loadDraft = async () => {
  const token = localStorage.getItem('AccessToken');
  try {
    const response = await fetch(`/api/drafts/${draftId.value}`, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
    if (response.ok) {
      const data = await response.json();
      briefTitle.value = data.title;
      versionHistory.value = data.backups || [];
      populateState(data.markdown);
    } else {
      errorMsg.value = 'Failed to load brief.';
    }
  } catch (err) {
    errorMsg.value = 'Network error loading brief.';
  }
};

const populateState = (markdown) => {
  const state = parseFromMarkdown(markdown);
  metadata.jurisdiction = state.metadata.jurisdiction;
  metadata.caseNumber = state.metadata.caseNumber;
  metadata.appellant = state.metadata.appellant;
  metadata.appellee = state.metadata.appellee;
  metadata.counselList = state.metadata.counselList || [];

  sections.a = state.sections.a;
  sections.d = state.sections.d;
  sections.e = state.sections.e;
  sections.f = state.sections.f;
  sections.g = state.sections.g;
  sections.h = state.sections.h;
  sections.i = state.sections.i;
  sections.j = state.sections.j;
  sections.k = state.sections.k;
};

// Live Cases Extractor
const liveCitations = computed(() => {
  const re = /\[\[(.*?)\]\]/g;
  const citations = new Set();
  
  // scan all edit fields
  const allTexts = [
    sections.a, sections.d, sections.e, sections.f,
    sections.g, sections.h, sections.i, sections.j, sections.k
  ].join('\n');

  let match;
  while ((match = re.exec(allTexts)) !== null) {
    if (match[1]) {
      const trimmed = match[1].trim();
      if (trimmed) {
        citations.add(trimmed);
      }
    }
  }

  const list = Array.from(citations);
  list.sort();
  return list;
});

const splitCitation = (citation) => {
  const parts = citation.split(/,(.+)/);
  if (parts.length > 1) {
    return { caseName: parts[0], remainder: ',' + parts[1] };
  }
  return { caseName: citation, remainder: '' };
};

// Counsel Management
const addCounsel = () => {
  metadata.counselList.push({
    name: '',
    firm: '',
    address: '',
    phone: '',
    email: ''
  });
};

const removeCounsel = (idx) => {
  metadata.counselList.splice(idx, 1);
};

// Save Draft
const saveBrief = async () => {
  isSaving.value = true;
  successMsg.value = '';
  errorMsg.value = '';

  const token = localStorage.getItem('AccessToken');
  const markdown = serializeToMarkdown();

  try {
    const response = await fetch('/api/drafts', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({
        id: draftId.value,
        title: briefTitle.value,
        markdown: markdown
      })
    });

    if (response.ok) {
      const data = await response.json();
      draftId.value = data.id;
      successMsg.value = 'Brief saved successfully! Created a revision backup.';
      // Reload draft details to refresh version list
      await loadDraft();
      
      // Update URL if we were in 'new' mode
      if (route.params.id === 'new') {
        router.replace(`/drafts/${data.id}`);
      }
    } else {
      errorMsg.value = 'Failed to save brief.';
    }
  } catch (err) {
    errorMsg.value = 'Network error saving brief.';
  } finally {
    isSaving.value = false;
    setTimeout(() => { successMsg.value = ''; }, 4000);
  }
};

// Restore revision backup
const restoreBackup = async (backupId) => {
  if (!confirm('Are you sure you want to restore this version? Your current unsaved edits will be overwritten.')) {
    return;
  }
  const token = localStorage.getItem('AccessToken');
  try {
    const response = await fetch(`/api/drafts/backups/${backupId}`, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
    if (response.ok) {
      const data = await response.json();
      populateState(data.markdown);
      successMsg.value = 'Restored previous revision to editor! Click Save to apply.';
      setTimeout(() => { successMsg.value = ''; }, 4000);
    } else {
      errorMsg.value = 'Failed to retrieve backup.';
    }
  } catch (e) {
    errorMsg.value = 'Network error.';
  }
};

// Import/Export Markdown
const triggerImport = () => {
  importFileInput.value.click();
};

const handleImportFile = (event) => {
  const file = event.target.files[0];
  if (!file) return;

  const reader = new FileReader();
  reader.onload = (e) => {
    const text = e.target.result;
    populateState(text);
    successMsg.value = 'Markdown brief imported successfully! Click Save to write to DB.';
    setTimeout(() => { successMsg.value = ''; }, 4000);
  };
  reader.readAsText(file);
};

const exportMarkdown = () => {
  const md = serializeToMarkdown();
  const blob = new Blob([md], { type: 'text/markdown;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  
  // sanitize filename
  const filename = briefTitle.value.toLowerCase().replace(/[^a-z0-9]/g, '_') + '.md';
  link.setAttribute('download', filename);
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
};

// PDF Download
const downloadPDF = () => {
  const token = localStorage.getItem('AccessToken');
  window.open(`/api/drafts/${draftId.value}/pdf?token=${token}`, '_blank');
};

const goBack = () => {
  router.push('/');
};

// Serialize State to Markdown
const serializeToMarkdown = () => {
  const fmData = {
    title: briefTitle.value,
    jurisdiction: metadata.jurisdiction,
    caseNumber: metadata.caseNumber,
    appellant: metadata.appellant,
    appellee: metadata.appellee,
    counselList: metadata.counselList
  };

  let md = `---json\n${JSON.stringify(fmData, null, 2)}\n---\n\n`;

  md += `# (a) Identity of Parties and Counsel\n${sections.a || ''}\n\n`;
  md += `# (d) Statement of the Case\n${sections.d || ''}\n\n`;
  md += `# (e) Any Statement Regarding Oral Argument\n${sections.e || ''}\n\n`;
  md += `# (f) Issues Presented\n${sections.f || ''}\n\n`;
  md += `# (g) Statement of Facts\n${sections.g || ''}\n\n`;
  md += `# (h) Summary of the Argument\n${sections.h || ''}\n\n`;
  md += `# (i) Argument\n${sections.i || ''}\n\n`;
  md += `# (j) Prayer\n${sections.j || ''}\n\n`;
  md += `# (k) Appendix\n${sections.k || ''}\n`;

  return md;
};

// Parse Markdown to state
const parseFromMarkdown = (md) => {
  const state = {
    title: 'Untitled Brief',
    metadata: {
      jurisdiction: '',
      caseNumber: '',
      appellant: '',
      appellee: '',
      counselList: []
    },
    sections: {
      a: '', d: '', e: '', f: '', g: '', h: '', i: '', j: '', k: ''
    }
  };

  if (!md) return state;

  let content = md;
  // Parse frontmatter
  if (md.startsWith('---json')) {
    const endIdx = md.indexOf('---', 7);
    if (endIdx !== -1) {
      try {
        const rawJson = md.substring(7, endIdx).trim();
        const fm = JSON.parse(rawJson);
        state.title = fm.title || 'Untitled Brief';
        state.metadata.jurisdiction = fm.jurisdiction || '';
        state.metadata.caseNumber = fm.caseNumber || '';
        state.metadata.appellant = fm.appellant || '';
        state.metadata.appellee = fm.appellee || '';
        state.metadata.counselList = fm.counselList || [];
      } catch (e) {
        console.error('Failed to parse JSON frontmatter:', e);
      }
      content = md.substring(endIdx + 3).trim();
    }
  } else if (md.startsWith('---')) {
    const endIdx = md.indexOf('---', 3);
    if (endIdx !== -1) {
      const rawFM = md.substring(3, endIdx).trim();
      const lines = rawFM.split('\n');
      for (const line of lines) {
        const colonIdx = line.indexOf(':');
        if (colonIdx !== -1) {
          const key = line.substring(0, colonIdx).trim();
          const val = line.substring(colonIdx + 1).trim().replace(/^["']|["']$/g, '');
          if (key === 'title') state.title = val;
          else if (key === 'jurisdiction') state.metadata.jurisdiction = val;
          else if (key === 'caseNumber') state.metadata.caseNumber = val;
          else if (key === 'appellant') state.metadata.appellant = val;
          else if (key === 'appellee') state.metadata.appellee = val;
        }
      }
      content = md.substring(endIdx + 3).trim();
    }
  }

  // Parse Sections
  const lines = content.split('\n');
  let currentSec = null;
  let secLines = [];

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const trimmed = line.trim();

    if (trimmed.startsWith('# (') && trimmed.length > 4 && trimmed[4] === ')') {
      if (currentSec && currentSec in state.sections) {
        state.sections[currentSec] = secLines.join('\n').trim();
      }
      currentSec = trimmed[3];
      secLines = [];
    } else {
      if (currentSec) {
        secLines.push(line);
      }
    }
  }
  if (currentSec && currentSec in state.sections) {
    state.sections[currentSec] = secLines.join('\n').trim();
  }

  return state;
};

// Smooth scroll to section
const scrollToSection = (id) => {
  const el = document.getElementById(id);
  if (el) {
    el.scrollIntoView({ behavior: 'smooth' });
    activeSection.value = id;
  }
};

// Detect active section on scroll
const onMainScroll = () => {
  const scroller = scrollerRef.value;
  if (!scroller) return;

  const scrollPos = scroller.scrollTop + 100; // offset

  for (const sec of tocSections) {
    const el = document.getElementById(sec.id);
    if (el) {
      const top = el.offsetTop;
      const height = el.offsetHeight;
      if (scrollPos >= top && scrollPos < top + height) {
        activeSection.value = sec.id;
        break;
      }
    }
  }
};

const formatDateTime = (dateStr) => {
  if (!dateStr) return '';
  const date = new Date(dateStr);
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
};
</script>

<style scoped>
.editor-layout {
  display: grid;
  grid-template-columns: 320px 1fr;
  height: 100vh;
  background-color: #f8fafc;
  overflow: hidden;
}

/* TOC Sidebar */
.toc-sidebar {
  background-color: #ffffff;
  border-right: 1px solid #e2e8f0;
  display: flex;
  flex-direction: column;
  padding: 20px;
  overflow-y: auto;
}

.back-link-wrapper {
  margin-bottom: 24px;
}

.btn-back {
  background: transparent;
  border: 1px solid #cbd5e1;
  color: #475569;
  border-radius: 8px;
  padding: 8px 14px;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.2s ease;
  width: 100%;
}

.btn-back:hover {
  background-color: #f1f5f9;
  color: #0f172a;
}

.btn-back svg {
  width: 16px;
  height: 16px;
}

.toc-header h3 {
  font-size: 0.75rem;
  font-weight: 800;
  color: #64748b;
  letter-spacing: 1.5px;
  margin-bottom: 12px;
  padding-left: 12px;
}

.toc-menu {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 24px;
}

.toc-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  color: #475569;
  text-decoration: none;
  font-size: 0.9rem;
  font-weight: 550;
  border-radius: 6px;
  transition: all 0.2s ease;
}

.toc-bullet {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background-color: #cbd5e1;
  transition: all 0.2s ease;
}

.toc-item:hover {
  color: #0f172a;
  background-color: #f1f5f9;
}

.toc-item.active {
  color: #1e3b8b;
  background-color: #eff6ff;
}

.toc-item.active .toc-bullet {
  background-color: #1e3b8b;
  transform: scale(1.5);
}

/* Citations Panel inside TOC */
.citations-panel {
  border-top: 1px solid #e2e8f0;
  padding-top: 20px;
  margin-top: auto;
}

.citations-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.citations-header h4 {
  font-size: 0.75rem;
  font-weight: 800;
  color: #64748b;
  letter-spacing: 1.5px;
  margin: 0;
}

.citations-count {
  background-color: #e2e8f0;
  color: #334155;
  font-size: 0.75rem;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 20px;
}

.citations-empty {
  font-size: 0.78rem;
  color: #94a3b8;
  line-height: 1.5;
  padding: 0 4px;
}

.citations-list {
  list-style: none;
  padding: 0;
  margin: 0;
  max-height: 180px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.citation-item {
  font-size: 0.8rem;
  color: #475569;
  line-height: 1.4;
  padding: 6px;
  background-color: #fafbfc;
  border-left: 3px solid #fbbf24;
  border-radius: 0 4px 4px 0;
}

/* Editor Main */
.editor-main {
  display: flex;
  flex-direction: column;
  height: 100vh;
}

.editor-header {
  background-color: #ffffff;
  border-bottom: 1px solid #e2e8f0;
  padding: 16px 32px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  z-index: 10;
}

.title-input-wrapper {
  flex-grow: 1;
}

.brief-title-input {
  width: 100%;
  font-size: 1.25rem;
  font-weight: 700;
  border: none;
  outline: none;
  color: #0f172a;
  padding: 6px 0;
  border-bottom: 1px solid transparent;
}

.brief-title-input:focus {
  border-color: #cbd5e1;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.btn-action-primary {
  padding: 10px 18px;
  background: linear-gradient(135deg, #1e3a8a 0%, #0f172a 100%);
  color: #ffffff;
  border: none;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-action-primary:hover:not(:disabled) {
  box-shadow: 0 4px 12px rgba(30, 58, 138, 0.2);
}

.btn-action-primary:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.btn-action-secondary {
  padding: 9px 14px;
  background-color: #ffffff;
  color: #475569;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all 0.2s ease;
}

.btn-action-secondary:hover {
  background-color: #f8fafc;
  color: #0f172a;
  border-color: #94a3b8;
}

.btn-action-secondary svg {
  width: 16px;
  height: 16px;
}

.btn-action-gold {
  padding: 9px 14px;
  background-color: #fef3c7;
  color: #92400e;
  border: 1px solid #fde68a;
  border-radius: 8px;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all 0.2s ease;
}

.btn-action-gold:hover:not(:disabled) {
  background-color: #fde68a;
  color: #78350f;
}

.btn-action-gold:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-action-gold svg {
  width: 16px;
  height: 16px;
}

.hidden-file-input {
  display: none;
}

/* Editor Scroll Container */
.editor-content-scroller {
  flex-grow: 1;
  overflow-y: auto;
  padding: 32px;
  display: flex;
  flex-direction: column;
  gap: 24px;
  scroll-behavior: smooth;
}

.success-banner {
  background-color: #ecfdf5;
  color: #065f46;
  border: 1px solid #a7f3d0;
  border-radius: 8px;
  padding: 12px 18px;
  font-size: 0.9rem;
  font-weight: 500;
}

.error-banner {
  background-color: #fef2f2;
  color: #991b1b;
  border: 1px solid #fca5a5;
  border-radius: 8px;
  padding: 12px 18px;
  font-size: 0.9rem;
  font-weight: 500;
}

/* Section Cards */
.editor-section-card {
  background-color: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 28px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.02);
}

.card-section-title {
  margin-bottom: 20px;
  border-bottom: 1px solid #f1f5f9;
  padding-bottom: 12px;
}

.card-section-title h2 {
  font-size: 1.15rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0 0 4px 0;
}

.info-tip {
  font-size: 0.8rem;
  color: #64748b;
  line-height: 1.4;
  display: inline-block;
}

/* Form Styling */
.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 24px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-group.full-width {
  grid-column: span 2;
}

.form-group label {
  font-size: 0.8rem;
  font-weight: 600;
  color: #475569;
}

.form-control {
  width: 100%;
  padding: 10px 14px;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  font-size: 0.92rem;
  transition: all 0.2s ease;
}

.form-control:focus {
  outline: none;
  border-color: #1e3a8a;
  box-shadow: 0 0 0 3px rgba(30, 58, 138, 0.1);
}

textarea.form-control {
  resize: vertical;
}

/* Textarea inputs for sections */
.section-textarea {
  width: 100%;
  min-height: 240px;
  padding: 16px;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  font-family: 'Georgia', serif;
  font-size: 1.05rem;
  line-height: 1.6;
  resize: vertical;
  transition: all 0.2s ease;
}

.section-textarea:focus {
  outline: none;
  border-color: #1e3a8a;
  box-shadow: 0 0 0 3px rgba(30, 58, 138, 0.1);
}

/* Counsel Manager */
.counsel-manager {
  border-top: 1px solid #edf2f7;
  padding-top: 20px;
  margin-top: 20px;
}

.counsel-manager-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.counsel-manager-header h3 {
  font-size: 0.95rem;
  font-weight: 700;
  color: #334155;
  margin: 0;
}

.btn-small-primary {
  padding: 6px 12px;
  background-color: #f1f5f9;
  color: #1e3a8a;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-small-primary:hover {
  background-color: #e2e8f0;
}

.counsel-empty {
  font-size: 0.85rem;
  color: #64748b;
  text-align: center;
  padding: 20px;
  background-color: #f8fafc;
  border-radius: 8px;
  border: 1px dashed #e2e8f0;
}

.counsel-grid-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.counsel-card {
  background-color: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 16px;
}

.counsel-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  border-bottom: 1px dashed #e2e8f0;
  padding-bottom: 8px;
}

.counsel-card-header h4 {
  font-size: 0.85rem;
  font-weight: 700;
  color: #475569;
  margin: 0;
}

.btn-remove-counsel {
  background: transparent;
  border: none;
  color: #e53e3e;
  font-size: 0.78rem;
  font-weight: 600;
  cursor: pointer;
}

.btn-remove-counsel:hover {
  text-decoration: underline;
}

.form-grid-mini {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.form-group-mini {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.form-group-mini.full-width {
  grid-column: span 2;
}

.form-group-mini label {
  font-size: 0.75rem;
  font-weight: 600;
  color: #64748b;
}

.form-control-mini {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 0.85rem;
}

.form-control-mini:focus {
  outline: none;
  border-color: #1e3a8a;
}

/* Read-Only Previews */
.read-only-section {
  background-color: #fafbfc;
  border-color: #e2e8f0;
}

.toc-preview-box, .ioa-preview-box {
  background-color: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 20px;
  font-family: 'Times New Roman', Times, serif;
}

.toc-preview-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.95rem;
  line-height: 1.8;
  margin-bottom: 4px;
}

.dot-leaders {
  flex-grow: 1;
  border-bottom: 1px dotted #94a3b8;
  margin: 0 8px;
  position: relative;
  top: 4px;
}

.no-cit-preview {
  font-family: sans-serif;
  font-size: 0.85rem;
  color: #64748b;
  line-height: 1.5;
  text-align: center;
  padding: 10px 0;
}

/* Revision History list */
.history-section {
  border-color: #cbd5e1;
}

.history-list-box {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.history-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background-color: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
}

.history-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.history-date {
  font-size: 0.85rem;
  font-weight: 700;
  color: #0f172a;
}

.history-title {
  font-size: 0.78rem;
  color: #64748b;
}

.btn-restore {
  padding: 6px 12px;
  background-color: #ffffff;
  border: 1px solid #cbd5e1;
  color: #1e3a8a;
  border-radius: 6px;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-restore:hover {
  background-color: #eff6ff;
  border-color: #93c5fd;
}
</style>
