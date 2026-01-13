// LocalCache - 本地缓存管理类
class LocalCache {
    constructor(ttlMinutes = 5) {
        this.ttl = ttlMinutes * 60 * 1000; // 转换为毫秒
        this.prefix = 'notex_cache_';
    }

    // 生成缓存键
    _makeKey(key) {
        return `${this.prefix}${key}`;
    }

    // 获取缓存
    get(key) {
        try {
            const fullKey = this._makeKey(key);
            const item = localStorage.getItem(fullKey);
            if (!item) return null;

            const data = JSON.parse(item);

            // 检查是否过期
            if (Date.now() > data.expiresAt) {
                localStorage.removeItem(fullKey);
                return null;
            }

            return data.value;
        } catch (e) {
            console.warn('Cache get error:', e);
            return null;
        }
    }

    // 设置缓存
    set(key, value, customTTL = null) {
        try {
            const fullKey = this._makeKey(key);
            const ttl = customTTL || this.ttl;
            const data = {
                value,
                expiresAt: Date.now() + ttl
            };
            localStorage.setItem(fullKey, JSON.stringify(data));
        } catch (e) {
            console.warn('Cache set error:', e);
        }
    }

    // 删除缓存
    delete(key) {
        try {
            const fullKey = this._makeKey(key);
            localStorage.removeItem(fullKey);
        } catch (e) {
            console.warn('Cache delete error:', e);
        }
    }

    // 按前缀删除缓存
    deletePattern(pattern) {
        try {
            const prefix = this._makeKey(pattern);
            const keys = [];
            for (let i = 0; i < localStorage.length; i++) {
                const key = localStorage.key(i);
                if (key && key.startsWith(prefix)) {
                    keys.push(key);
                }
            }
            keys.forEach(key => localStorage.removeItem(key));
        } catch (e) {
            console.warn('Cache deletePattern error:', e);
        }
    }

    // 清空所有缓存
    clear() {
        try {
            const keys = [];
            for (let i = 0; i < localStorage.length; i++) {
                const key = localStorage.key(i);
                if (key && key.startsWith(this.prefix)) {
                    keys.push(key);
                }
            }
            keys.forEach(key => localStorage.removeItem(key));
        } catch (e) {
            console.warn('Cache clear error:', e);
        }
    }

    // 清理过期缓存
    cleanup() {
        try {
            const now = Date.now();
            const keys = [];
            for (let i = 0; i < localStorage.length; i++) {
                const key = localStorage.key(i);
                if (key && key.startsWith(this.prefix)) {
                    keys.push(key);
                }
            }
            keys.forEach(key => {
                try {
                    const item = localStorage.getItem(key);
                    if (item) {
                        const data = JSON.parse(item);
                        if (now > data.expiresAt) {
                            localStorage.removeItem(key);
                        }
                    }
                } catch (e) {
                    // 忽略解析错误，删除无效条目
                    localStorage.removeItem(key);
                }
            });
        } catch (e) {
            console.warn('Cache cleanup error:', e);
        }
    }
}

class OpenNotebook {
    constructor() {
        this.notebooks = [];
        this.currentNotebook = null;
        this.apiBase = '/api';
        this.currentChatSession = null;
        this.config = {
            allowDelete: true
        };

        // 初始化本地缓存 (5分钟TTL)
        this.cache = new LocalCache(5);

        // Note type name mapping
        this.noteTypeNameMap = {
            summary: '摘要',
            faq: '常见问题',
            study_guide: '学习指南',
            outline: '大纲',
            podcast: '播客',
            timeline: '时间线',
            glossary: '术语表',
            quiz: '测验',
            mindmap: '思维导图',
            infograph: '信息图',
            ppt: '幻灯片',
            insight: '洞察报告'
        };

        this.init();
    }

    async init() {
        await this.loadConfig();
        this.bindEvents();
        this.initResizers();
        this.switchView('landing');

        // 清理过期缓存
        this.cache.cleanup();

        await this.loadNotebooks();
        this.applyConfig();
    }

    async loadConfig() {
        try {
            const config = await this.api('/config');
            // Convert snake_case to camelCase
            this.config = {
                allowDelete: config.allow_delete !== undefined ? config.allow_delete : true
            };
        } catch (error) {
            console.error('Failed to load config:', error);
        }
    }

    applyConfig() {
        // Show/hide delete buttons based on allowDelete config
        const deleteButtons = document.querySelectorAll('.btn-delete, .btn-remove-source, .btn-delete-note, .btn-delete-card, .btn-delete-notebook, .btn-delete-compact-note');
        deleteButtons.forEach(btn => {
            if (this.config.allowDelete) {
                btn.style.display = '';
                btn.style.opacity = '1';
            } else {
                btn.style.display = 'none';
            }
        });
    }

    initResizers() {
        const resizerLeft = document.getElementById('resizerLeft');
        const resizerRight = document.getElementById('resizerRight');
        const grid = document.querySelector('.main-grid');

        if (!resizerLeft || !resizerRight) return;

        let isDragging = false;
        let currentResizer = null;

        const startDragging = (e, resizer) => {
            isDragging = true;
            currentResizer = resizer;
            resizer.classList.add('dragging');
            document.body.style.cursor = 'col-resize';
            e.preventDefault();
        };

        const stopDragging = () => {
            if (!isDragging) return;
            isDragging = false;
            currentResizer.classList.remove('dragging');
            document.body.style.cursor = '';
            currentResizer = null;
        };

        const drag = (e) => {
            if (!isDragging) return;

            const gridRect = grid.getBoundingClientRect();
            if (currentResizer === resizerLeft) {
                const width = e.clientX - gridRect.left;
                if (width > 150 && width < 600) {
                    grid.style.setProperty('--left-width', `${width}px`);
                }
            } else if (currentResizer === resizerRight) {
                const width = gridRect.right - e.clientX;
                if (width > 200 && width < 600) {
                    grid.style.setProperty('--right-width', `${width}px`);
                }
            }
        };

        resizerLeft.addEventListener('mousedown', (e) => startDragging(e, resizerLeft));
        resizerRight.addEventListener('mousedown', (e) => startDragging(e, resizerRight));
        document.addEventListener('mousemove', drag);
        document.addEventListener('mouseup', stopDragging);
    }

    bindEvents() {
        const safeAddEventListener = (id, event, handler) => {
            const el = document.getElementById(id);
            if (el) el.addEventListener(event, handler);
        };

        safeAddEventListener('btnNewNotebook', 'click', () => this.showNewNotebookModal());
        safeAddEventListener('btnNewNotebookLanding', 'click', () => this.showNewNotebookModal());
        safeAddEventListener('btnBackToList', 'click', () => this.switchView('landing'));
        safeAddEventListener('btnToggleRight', 'click', () => this.toggleRightPanel());
        safeAddEventListener('btnToggleLeft', 'click', () => this.toggleLeftPanel());
        safeAddEventListener('btnShowNotesDetails', 'click', () => this.showNotesListTab());
        safeAddEventListener('btnCloseNotesList', 'click', (e) => {
            e.stopPropagation();
            this.closeNotesListTab();
        });
        safeAddEventListener('btnCloseNote', 'click', (e) => {
            e.stopPropagation();
            this.closeNoteTab();
        });

        // Panel tabs
        document.querySelectorAll('.tab-btn').forEach(tab => {
            tab.addEventListener('click', () => {
                this.switchPanelTab(tab.dataset.tab);
            });
        });
        
        safeAddEventListener('newNotebookForm', 'submit', (e) => this.handleCreateNotebook(e));
        safeAddEventListener('btnCloseNotebookModal', 'click', () => this.closeModals());
        safeAddEventListener('btnCancelNotebook', 'click', () => this.closeModals());

        safeAddEventListener('btnAddSource', 'click', () => this.showAddSourceModal());
        safeAddEventListener('btnCloseSourceModal', 'click', () => this.closeModals());
        const dropZone = document.getElementById('dropZone');
        if (dropZone) {
            dropZone.addEventListener('click', () => document.getElementById('fileInput').click());
            dropZone.addEventListener('dragover', (e) => {
                e.preventDefault();
                dropZone.classList.add('drag-over');
            });
            dropZone.addEventListener('dragleave', () => {
                dropZone.classList.remove('drag-over');
            });
            dropZone.addEventListener('drop', (e) => this.handleDrop(e));
        }
        
        safeAddEventListener('fileInput', 'change', (e) => this.handleFileUpload(e));
        safeAddEventListener('textSourceForm', 'submit', (e) => this.handleTextSource(e));
        safeAddEventListener('urlSourceForm', 'submit', (e) => this.handleURLSource(e));
        safeAddEventListener('btnCancelText', 'click', () => this.closeModals());
        safeAddEventListener('btnCancelURL', 'click', () => this.closeModals());

        document.querySelectorAll('.source-tab').forEach(tab => {
            tab.addEventListener('click', () => {
                document.querySelectorAll('.source-tab').forEach(t => t.classList.remove('active'));
                document.querySelectorAll('.source-content').forEach(c => c.classList.remove('active'));
                tab.classList.add('active');
                const targetId = `source${tab.dataset.source.charAt(0).toUpperCase() + tab.dataset.source.slice(1)}`;
                const target = document.getElementById(targetId);
                if (target) target.classList.add('active');
            });
        });

        document.querySelectorAll('.transform-card').forEach(card => {
            card.addEventListener('click', (e) => {
                e.preventDefault();
                this.handleTransform(card.dataset.type, card);
            });
        });

        safeAddEventListener('btnCustomTransform', 'click', (e) => {
            this.handleTransform('custom', e.currentTarget);
        });

        safeAddEventListener('chatForm', 'submit', (e) => this.handleChat(e));

        safeAddEventListener('modalOverlay', 'click', (e) => {
            if (e.target.id === 'modalOverlay') {
                this.closeModals();
            }
        });
    }

    // API 方法
    async api(endpoint, options = {}) {
        const timeout = options.timeout || 300000; // 默认 300 秒
        const controller = new AbortController();
        const id = setTimeout(() => controller.abort(), timeout);

        const defaults = {
            headers: {
                'Content-Type': 'application/json',
            },
            cache: 'no-store',
            signal: controller.signal
        };

        let url = `${this.apiBase}${endpoint}`;
        if (!options.method || options.method === 'GET') {
            const separator = url.includes('?') ? '&' : '?';
            url += `${separator}_t=${Date.now()}`;
        }

        try {
            const response = await fetch(url, { ...defaults, ...options });
            clearTimeout(id);

            if (!response.ok) {
                const error = await response.json().catch(() => ({ error: '请求失败' }));
                throw new Error(error.error || '请求失败');
            }

            if (response.status === 204) {
                return null;
            }

            return response.json();
        } catch (error) {
            clearTimeout(id);
            if (error.name === 'AbortError') {
                throw new Error('请求超时，请稍后重试');
            }
            throw error;
        }
    }

    getWebSocketURL(path) {
        const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
        return `${protocol}://${window.location.host}${path}`;
    }

    // 笔记本方法
    async loadNotebooks() {
        try {
            // 先尝试从缓存获取
            const cached = this.cache.get('notebooks');
            if (cached) {
                this.notebooks = cached;
                this.renderNotebooks();
                this.updateFooter();
            }

            // 从服务器获取最新数据（包含统计信息）
            const notebooks = await this.api('/notebooks/stats');
            this.notebooks = notebooks;

            // 更新缓存
            this.cache.set('notebooks', notebooks);

            this.renderNotebooks();
            this.updateFooter();
        } catch (error) {
            this.showError('加载笔记本失败');
        }
    }

    renderNotebooks() {
        this.renderNotebookCards();
    }

    renderNotebookCards() {
        const container = document.getElementById('notebookGridLanding');
        const template = document.getElementById('notebookCardTemplate');

        container.innerHTML = '';

        if (this.notebooks.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <svg width="64" height="64" viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="1">
                        <rect x="12" y="12" width="40" height="40" rx="4"/>
                        <line x1="20" y1="24" x2="44" y2="24"/>
                        <line x1="20" y1="32" x2="40" y2="32"/>
                    </svg>
                    <p>开启你的知识之旅</p>
                    <button class="btn-primary" onclick="app.showNewNotebookModal()">创建第一个笔记本</button>
                </div>
            `;
            return;
        }

        this.notebooks.forEach(nb => {
            const clone = template.content.cloneNode(true);
            const card = clone.querySelector('.notebook-card');

            card.dataset.id = nb.id;
            card.querySelector('.notebook-card-name').textContent = nb.name;
            card.querySelector('.notebook-card-desc').textContent = nb.description || '暂无描述';

            // 直接使用从 API 获取的统计信息
            card.querySelector('.stat-sources').textContent = `${nb.source_count || 0} 来源`;
            card.querySelector('.stat-notes').textContent = `${nb.note_count || 0} 笔记`;

            card.addEventListener('click', (e) => {
                if (!e.target.closest('.btn-delete-card')) {
                    this.selectNotebook(nb.id);
                }
            });

            const deleteCardBtn = card.querySelector('.btn-delete-card');
            deleteCardBtn.addEventListener('click', (e) => {
                e.stopPropagation();
                if (confirm('确定要删除此笔记本吗？')) {
                    this.deleteNotebook(nb.id);
                }
            });

            // Apply config to delete button
            if (this.config.allowDelete) {
                deleteCardBtn.style.display = 'flex';
            } else {
                deleteCardBtn.style.display = 'none';
            }

            container.appendChild(clone);
        });
    }

    switchView(view) {
        const landing = document.getElementById('landingPage');
        const workspace = document.getElementById('workspaceContainer');
        const header = document.querySelector('.app-header');

        if (view === 'workspace') {
            landing.classList.add('hidden');
            workspace.classList.remove('hidden');
            header.classList.add('hidden');
        } else {
            landing.classList.remove('hidden');
            workspace.classList.add('hidden');
            header.classList.remove('hidden');
            this.currentNotebook = null;
            this.renderNotebookCards();
        }
    }

    toggleRightPanel() {
        const grid = document.querySelector('.main-grid');
        grid.classList.toggle('right-collapsed');
    }

    toggleLeftPanel() {
        const grid = document.querySelector('.main-grid');
        grid.classList.toggle('left-collapsed');
    }

    switchPanelTab(tab) {
        // Update tab buttons
        document.querySelectorAll('.tab-btn').forEach(t => {
            t.classList.toggle('active', t.dataset.tab === tab);
        });

        // Update content visibility
        const chatWrapper = document.querySelector('.chat-messages-wrapper');
        const noteViewContainer = document.querySelector('.note-view-container');
        const notesDetailsView = document.querySelector('.notes-details-view');

        if (tab === 'note') {
            chatWrapper.style.display = 'none';
            if (notesDetailsView) notesDetailsView.style.display = 'none';
            if (noteViewContainer) {
                noteViewContainer.style.display = 'flex';
            }
        } else if (tab === 'chat') {
            chatWrapper.style.display = 'flex';
            if (notesDetailsView) notesDetailsView.style.display = 'none';
            if (noteViewContainer) {
                noteViewContainer.style.display = 'none';
            }
        } else if (tab === 'notes_list') {
            chatWrapper.style.display = 'none';
            if (noteViewContainer) noteViewContainer.style.display = 'none';
            if (notesDetailsView) {
                notesDetailsView.style.display = 'flex';
                this.renderNotesCompactGrid();
            }
        }
    }

    async showNotesListTab() {
        const tabBtn = document.getElementById('tabBtnNotesList');
        tabBtn.classList.remove('hidden');

        // Ensure notesDetailsView container exists
        let notesDetailsView = document.querySelector('.notes-details-view');
        if (!notesDetailsView) {
            const chatWrapper = document.querySelector('.chat-messages-wrapper');
            notesDetailsView = document.createElement('div');
            notesDetailsView.className = 'notes-details-view';
            notesDetailsView.innerHTML = '<div class="notes-compact-grid"></div>';
            chatWrapper.insertAdjacentElement('afterend', notesDetailsView);
        }

        this.switchPanelTab('notes_list');
    }

    closeNotesListTab() {
        const tabBtn = document.getElementById('tabBtnNotesList');
        tabBtn.classList.add('hidden');
        
        const notesDetailsView = document.querySelector('.notes-details-view');
        if (notesDetailsView) notesDetailsView.style.display = 'none';
        
        if (tabBtn.classList.contains('active')) {
            this.switchPanelTab('chat');
        }
    }

    closeNoteTab() {
        const noteViewContainer = document.querySelector('.note-view-container');
        if (noteViewContainer) noteViewContainer.remove();
        
        const tabBtnNote = document.getElementById('tabBtnNote');
        if (tabBtnNote) tabBtnNote.style.display = 'none';

        this.switchPanelTab('chat');
    }

    async renderNotesCompactGrid() {
        if (!this.currentNotebook) return;

        const container = document.querySelector('.notes-compact-grid');
        if (!container) return;

        try {
            const notes = await this.api(`/notebooks/${this.currentNotebook.id}/notes`);
            container.innerHTML = '';

            notes.forEach(note => {
                const card = document.createElement('div');
                card.className = 'compact-note-card';

                const plainText = note.content
                    .replace(/^#+\s+/gm, '')
                    .replace(/\*\*/g, '')
                    .replace(/\*/g, '')
                    .replace(/`/g, '')
                    .replace(/\n+/g, ' ')
                    .trim();

                card.innerHTML = `
                    <button class="btn-delete-compact-note" title="删除笔记">
                        <svg width="14" height="14" viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5">
                            <polyline points="4,3 4,2 10,2 10,3"/>
                            <line x1="2" y1="4" x2="12" y2="4"/>
                            <line x1="5" y1="7" x2="5" y2="11"/>
                            <line x1="9" y1="7" x2="9" y2="11"/>
                        </svg>
                    </button>
                    <div class="note-type">${note.type}</div>
                    <h4 class="note-title">${note.title}</h4>
                    <p class="note-preview">${plainText}</p>
                    <div class="note-footer">
                        <span>${this.formatDate(note.created_at)}</span>
                        <span>${note.source_ids?.length || 0} 来源</span>
                    </div>
                `;

                // Delete button event
                const deleteBtn = card.querySelector('.btn-delete-compact-note');
                deleteBtn.addEventListener('click', (e) => {
                    e.stopPropagation();
                    if (confirm('确定要删除此笔记吗？')) {
                        this.deleteNote(note.id);
                    }
                });

                // Apply config to delete button
                if (this.config.allowDelete) {
                    deleteBtn.style.display = 'flex';
                } else {
                    deleteBtn.style.display = 'none';
                }

                card.addEventListener('click', () => this.viewNote(note));
                container.appendChild(card);
            });
        } catch (error) {
            console.error('Failed to load notes for grid:', error);
        }
    }

    async selectNotebook(id) {
        this.currentNotebook = this.notebooks.find(nb => nb.id === id);
        
        document.getElementById('currentNotebookName').textContent = this.currentNotebook.name;
        this.switchView('workspace');
        
        // Reset tab to notes list and remove any existing note view
        this.showNotesListTab();
        const noteView = document.querySelector('.note-view-container');
        if (noteView) noteView.remove();

        await Promise.all([
            this.loadSources(),
            this.loadNotes(),
            this.loadChatSessions()
        ]);

        this.setStatus(`当前选择: ${this.currentNotebook.name}`);
    }

    showNewNotebookModal() {
        document.getElementById('newNotebookModal').classList.add('active');
        document.getElementById('modalOverlay').classList.add('active');
        document.querySelector('#newNotebookForm input[name="name"]').focus();
    }

    async handleCreateNotebook(e) {
        e.preventDefault();
        const form = e.target;
        const data = new FormData(form);

        this.showLoading('处理中...');

        try {
            const notebook = await this.api('/notebooks', {
                method: 'POST',
                body: JSON.stringify({
                    name: data.get('name'),
                    description: data.get('description') || undefined,
                }),
            });

            // 使缓存失效
            this.cache.delete('notebooks');

            this.notebooks.push(notebook);
            this.renderNotebooks();
            this.selectNotebook(notebook.id);
            this.closeModals();
            form.reset();
            this.hideLoading();
        } catch (error) {
            this.hideLoading();
            this.showError(error.message);
        }
    }

    async deleteNotebook(id) {
        try {
            await this.api(`/notebooks/${id}`, { method: 'DELETE' });

            // 使缓存失效
            this.cache.delete('notebooks');
            this.cache.deletePattern(`sources_${id}`);
            this.cache.deletePattern(`notes_${id}`);
            this.cache.deletePattern(`chat_${id}`);

            this.notebooks = this.notebooks.filter(nb => nb.id !== id);

            if (this.currentNotebook?.id === id) {
                this.currentNotebook = null;
                this.clearContentAreas();
                this.switchView('landing');
            }

            this.renderNotebooks();
            this.updateFooter();
        } catch (error) {
            this.showError('删除笔记本失败: ' + error.message);
        }
    }

    clearContentAreas() {
        const sourcesContainer = document.getElementById('sourcesGrid');
        sourcesContainer.innerHTML = `
            <div class="empty-state">
                <svg width="64" height="64" viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="1">
                    <path d="M20 8 L44 8 L48 12 L48 56 L20 56 Z"/>
                    <polyline points="44,8 44,12 48,12"/>
                    <line x1="28" y1="24" x2="40" y2="24"/>
                    <line x1="28" y1="32" x2="40" y2="32"/>
                    <line x1="28" y1="40" x2="36" y2="40"/>
                </svg>
                <p>添加来源以开始</p>
                <p class="empty-hint">支持 PDF, TXT, MD, DOCX, HTML</p>
            </div>
        `;

        const notesContainer = document.getElementById('notesList');
        notesContainer.innerHTML = `
            <div class="empty-state">
                <svg width="48" height="48" viewBox="0 0 48 48" fill="none" stroke="currentColor" stroke-width="1.5">
                    <path d="M12 4 L36 4 L40 8 L40 44 L12 44 Z"/>
                    <polyline points="36,4 36,8 40,8"/>
                </svg>
                <p>暂无笔记</p>
                <p class="empty-hint">使用转换从来源生成笔记</p>
            </div>
        `;

        const chatContainer = document.getElementById('chatMessages');
        chatContainer.innerHTML = `
            <div class="chat-welcome">
                <svg width="40" height="40" viewBox="0 0 40 40" fill="none" stroke="currentColor" stroke-width="1.5">
                    <circle cx="20" cy="12" r="6"/>
                    <path d="M8 38 C8 28 14 22 20 22 C26 22 32 28 32 38"/>
                </svg>
                <h3>与来源对话</h3>
                <p>询问关于笔记本内容的问题</p>
            </div>
        `;

        this.currentChatSession = null;
    }

    // 来源方法
    async loadSources() {
        if (!this.currentNotebook) return;

        const container = document.getElementById('sourcesGrid');
        const template = document.getElementById('sourceTemplate');

        try {
            // 先尝试从缓存获取
            const cacheKey = `sources_${this.currentNotebook.id}`;
            const cached = this.cache.get(cacheKey);

            // 从服务器获取最新数据
            const sources = await this.api(`/notebooks/${this.currentNotebook.id}/sources`);

            // 更新缓存
            this.cache.set(cacheKey, sources);

            if (sources.length === 0) {
                this.clearContentAreas();
                return;
            }

            container.innerHTML = '';

            sources.forEach(source => {
                const clone = template.content.cloneNode(true);
                const card = clone.querySelector('.source-card');

                card.dataset.id = source.id;
                card.querySelector('.source-type-badge').textContent = source.type;
                card.querySelector('.source-name').textContent = source.name;
                card.querySelector('.source-meta').textContent = this.formatFileSize(source.file_size) || '文本来源';
                card.querySelector('.chunk-count').textContent = source.chunk_count || 0;

                const icon = this.getSourceIcon(source.type);
                card.querySelector('.source-icon').innerHTML = icon;

                const removeBtn = card.querySelector('.btn-remove-source');
                removeBtn.addEventListener('click', () => {
                    this.removeSource(source.id);
                });

                // Apply config to delete button
                if (this.config.allowDelete) {
                    removeBtn.style.display = 'flex';
                } else {
                    removeBtn.style.display = 'none';
                }

                container.appendChild(clone);
            });

            this.updateFooter();
        } catch (error) {
            console.error('加载来源失败:', error);
        }
    }

    getSourceIcon(type) {
        const icons = {
            file: '<svg viewBox="0 0 40 40" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M10 4 L24 4 L30 10 L30 36 L10 36 Z"/><polyline points="24,4 24,10 30,10"/></svg>',
            text: '<svg viewBox="0 0 40 40" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M8 6 L32 6"/><path d="M8 12 L32 12"/><path d="M8 18 L28 18"/><path d="M8 24 L32 24"/><path d="M8 30 L24 30"/></svg>',
            url: '<svg viewBox="0 0 40 40" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M12 20 C12 14 16 10 22 10 C28 10 32 14 32 20 C32 26 28 30 22 30"/><path d="M28 20 C28 26 24 30 18 30 C12 30 8 26 8 20 C8 14 12 10 18 10"/></svg>',
            insight: '<svg viewBox="0 0 40 40" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="20" cy="20" r="14"/><path d="M20 12 L20 22"/><path d="M20 26 L20 28"/><circle cx="20" cy="20" r="8" stroke-dasharray="2 2"/></svg>',
        };
        return icons[type] || icons.file;
    }

    formatFileSize(bytes) {
        if (!bytes) return null;
        if (bytes < 1024) return bytes + ' B';
        if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
        return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
    }

    showAddSourceModal() {
        if (!this.currentNotebook) {
            this.showError('请先选择一个笔记本');
            return;
        }
        document.getElementById('addSourceModal').classList.add('active');
        document.getElementById('modalOverlay').classList.add('active');
    }

    async handleFileUpload(e) {
        const files = e.target.files;
        if (!files.length) return;

        this.showLoading('处理中...');

        for (const file of files) {
            const formData = new FormData();
            formData.append('file', file);
            formData.append('notebook_id', this.currentNotebook.id);

            try {
                await this.api('/upload', {
                    method: 'POST',
                    headers: {},
                    body: formData,
                });
            } catch (error) {
                this.showError(`上传失败: ${file.name} - ${error.message}`);
            }
        }

        this.hideLoading();
        this.closeModals();
        await this.loadSources();
        await this.updateCurrentNotebookCounts();
        document.getElementById('fileInput').value = '';
    }

    async handleTextSource(e) {
        e.preventDefault();
        const form = e.target;
        const data = new FormData(form);

        this.showLoading('处理中...');

        try {
            await this.api(`/notebooks/${this.currentNotebook.id}/sources`, {
                method: 'POST',
                body: JSON.stringify({
                    name: data.get('name'),
                    type: 'text',
                    content: data.get('content'),
                }),
            });

            this.hideLoading();
            this.closeModals();
            form.reset();
            await this.loadSources();
            await this.updateCurrentNotebookCounts();
        } catch (error) {
            this.hideLoading();
            this.showError(error.message);
        }
    }

    async handleURLSource(e) {
        e.preventDefault();
        const form = e.target;
        const data = new FormData(form);

        this.showLoading('获取网址内容中...');

        try {
            await this.api(`/notebooks/${this.currentNotebook.id}/sources`, {
                method: 'POST',
                body: JSON.stringify({
                    name: data.get('name') || data.get('url'),
                    type: 'url',
                    url: data.get('url'),
                }),
            });

            this.hideLoading();
            this.closeModals();
            form.reset();
            await this.loadSources();
            await this.updateCurrentNotebookCounts();
        } catch (error) {
            this.hideLoading();
            this.showError(error.message);
        }
    }

    handleDrop(e) {
        e.preventDefault();
        document.getElementById('dropZone').classList.remove('drag-over');

        const files = e.dataTransfer.files;
        if (!files.length) return;

        document.getElementById('fileInput').files = files;
        this.handleFileUpload({ target: { files } });
    }

    async removeSource(id) {
        try {
            await this.api(`/notebooks/${this.currentNotebook.id}/sources/${id}`, {
                method: 'DELETE',
            });
            await this.loadSources();
            await this.updateCurrentNotebookCounts();
        } catch (error) {
            this.showError('移除来源失败');
        }
    }

    // 笔记方法
    async loadNotes() {
        if (!this.currentNotebook) return;

        const container = document.getElementById('notesList');
        const template = document.getElementById('noteTemplate');
        const countHeader = document.querySelector('.section-notes .panel-title');

        try {
            // 先尝试从缓存获取
            const cacheKey = `notes_${this.currentNotebook.id}`;
            const cached = this.cache.get(cacheKey);

            // 从服务器获取最新数据
            const notes = await this.api(`/notebooks/${this.currentNotebook.id}/notes`);

            // 更新缓存
            this.cache.set(cacheKey, notes);
            
            if (countHeader) {
                countHeader.textContent = `笔记 (${notes.length})`;
            }

            if (notes.length === 0) {
                container.innerHTML = `
                    <div class="empty-state">
                        <svg width="48" height="48" viewBox="0 0 48 48" fill="none" stroke="currentColor" stroke-width="1.5">
                            <path d="M12 4 L36 4 L40 8 L40 44 L12 44 Z"/>
                            <polyline points="36,4 36,8 40,8"/>
                        </svg>
                        <p>暂无笔记</p>
                        <p class="empty-hint">使用转换从来源生成笔记</p>
                    </div>
                `;
                return;
            }

            container.innerHTML = '';

            notes.forEach(note => {
                const clone = template.content.cloneNode(true);
                const item = clone.querySelector('.note-item');

                item.dataset.id = note.id;
                item.querySelector('.note-type-badge').textContent = this.noteTypeNameMap[note.type] || note.type.toUpperCase();
                item.querySelector('.note-title').textContent = note.title;

                const plainText = note.content
                    .replace(/^#+\s+/gm, '')
                    .replace(/\*\*/g, '')
                    .replace(/\*/g, '')
                    .replace(/`/g, '')
                    .replace(/\ \[([^\]]+)\]\([^)]+\)/g, '$1')
                    .replace(/\n+/g, ' ')
                    .trim();

                item.querySelector('.note-preview').textContent = plainText;
                item.querySelector('.note-date').textContent = this.formatDate(note.created_at);
                item.querySelector('.note-sources').textContent = `${note.source_ids?.length || 0} 来源`;

                const deleteBtn = item.querySelector('.btn-delete-note');
                deleteBtn.addEventListener('click', () => {
                    this.deleteNote(note.id);
                });

                // Apply config to delete button
                if (this.config.allowDelete) {
                    deleteBtn.style.display = 'flex';
                } else {
                    deleteBtn.style.display = 'none';
                }

                item.addEventListener('click', (e) => {
                    if (!e.target.closest('.btn-delete-note')) {
                        this.viewNote(note);
                    }
                });

                container.appendChild(clone);
            });

            this.updateFooter();
        } catch (error) {
            console.error('加载笔记失败:', error);
        }
    }

    async viewNote(note) {
        const renderedContent = marked.parse(note.content);
        const infographicHTML = note.metadata?.image_url 
            ? `<div class="infographic-container">
                 <img src="${note.metadata.image_url}" alt="Infographic" class="infographic-image">
                 <div class="infographic-actions">
                    <a href="${note.metadata.image_url}" target="_blank" class="btn-text">查看大图</a>
                 </div>
               </div>`
            : '';

        // PPT Slider HTML
        let pptSliderHTML = '';
        if (note.metadata?.slides && note.metadata.slides.length > 0) {
            const slides = note.metadata.slides;
            pptSliderHTML = `
                <div class="ppt-viewer-container" id="pptViewer">
                    <div class="ppt-slides-wrapper">
                        ${slides.map((src, index) => `
                            <div class="ppt-slide ${index === 0 ? 'active' : ''}" data-index="${index}">
                                <img src="${src}" alt="Slide ${index + 1}">
                                <div class="ppt-slide-counter">${index + 1} / ${slides.length}</div>
                            </div>
                        `).join('')}
                    </div>
                    <div class="ppt-controls">
                        <button class="btn-ppt-nav prev" id="btnPptPrev">
                            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="15 18 9 12 15 6"></polyline></svg>
                        </button>
                        <button class="btn-ppt-nav next" id="btnPptNext">
                            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"></polyline></svg>
                        </button>
                    </div>
                </div>
            `;
        }

        // Determine if we should show the text content
        const showMarkdownContent = (note.type !== 'infograph' && note.type !== 'ppt') || (!note.metadata?.image_url && !note.metadata?.slides);

        // Show the Note tab button
        const tabBtnNote = document.getElementById('tabBtnNote');
        if (tabBtnNote) {
            tabBtnNote.style.display = 'flex';
        }

        // Remove existing note view if any
        const existingNoteView = document.querySelector('.note-view-container');
        if (existingNoteView) {
            existingNoteView.remove();
        }

        // Create note view container and insert it after chat-messages-wrapper
        const noteViewHTML = `
            <div class="note-view-container">
                <div class="note-view-header">
                    <div class="note-view-info">
                        <span class="note-view-type">${note.type}</span>
                        <span class="note-view-title-text">${note.title}</span>
                    </div>
                    <div class="note-view-actions">
                        <button class="btn-copy-note" id="btnCopyNote" title="复制 Markdown">
                            <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2">
                                <rect x="3" y="3" width="10" height="10" rx="1"/>
                                <path d="M7 3 L7 1 C7 1 13 1 13 1 L13 13 L11 13"/>
                            </svg>
                        </button>
                    </div>
                </div>
                <div class="note-view-content">
                    ${infographicHTML}
                    ${pptSliderHTML}
                    <div class="markdown-content" style="${showMarkdownContent ? '' : 'display:none'}">${renderedContent}</div>
                </div>
            </div>
        `;

        const chatWrapper = document.querySelector('.chat-messages-wrapper');
        chatWrapper.insertAdjacentHTML('afterend', noteViewHTML);

        // PPT Navigation Logic
        if (note.metadata?.slides) {
            let currentSlide = 0;
            const slidesCount = note.metadata.slides.length;
            const slideElements = document.querySelectorAll('.ppt-slide');
            
            const showSlide = (n) => {
                slideElements[currentSlide].classList.remove('active');
                currentSlide = (n + slidesCount) % slidesCount;
                slideElements[currentSlide].classList.add('active');
            };

            document.getElementById('btnPptPrev').addEventListener('click', () => showSlide(currentSlide - 1));
            document.getElementById('btnPptNext').addEventListener('click', () => showSlide(currentSlide + 1));
            
            // Key navigation
            const keyHandler = (e) => {
                if (e.key === 'ArrowLeft') showSlide(currentSlide - 1);
                if (e.key === 'ArrowRight') showSlide(currentSlide + 1);
            };
            document.addEventListener('keydown', keyHandler);
            // Cleanup handler on container remove? We'll leave it for now or add observer
        }

        // Render Mermaid diagrams if any
        if (window.mermaid) {
            try {
                mermaid.initialize({ 
                    startOnLoad: false, 
                    theme: 'base',
                    securityLevel: 'loose',
                    fontFamily: 'var(--font-sans)',
                    themeVariables: {
                        // Vibrant WeChat Green Theme
                        primaryColor: '#ecfdf5', // Lighter, more vibrant green background
                        primaryTextColor: '#065f46', // Deep emerald for text
                        primaryBorderColor: '#10b981', // Bright emerald border
                        lineColor: '#10b981', // Bright line color
                        secondaryColor: '#f0fdf4',
                        tertiaryColor: '#ffffff',
                        fontSize: '14px',
                        mainBkg: '#ecfdf5',
                        nodeBorder: '#10b981',
                        clusterBkg: '#f0fdf4',
                        // Mindmap specific vibrancy
                        nodeTextColor: '#065f46',
                        edgeColor: '#34d399' // Slightly lighter green for edges
                    },
                    mindmap: {
                        useMaxWidth: true,
                        padding: 20
                    }
                });
                
                const contentArea = document.querySelector('.note-view-content');
                const mermaidBlocks = contentArea.querySelectorAll('pre code.language-mermaid');
                
                // Helper to fix common mermaid errors
                const sanitizeMermaid = (code) => {
                    let sanitized = code.trim();

                    // 1. If it's a graph and has unquoted brackets, try to wrap them
                    if (sanitized.startsWith('graph')) {
                        // Fix things like: A --> socket() --> B
                        sanitized = sanitized.replace(/(\s+)-->(\s+)([^"\s][^-\n>]*\([^)]*\)[^-\n>]*)/g, '$1-->$2"$3"');
                        sanitized = sanitized.replace(/([^"\s][^-\n>]*\([^)]*\)[^-\n>]*)\s+-->/g, '"$1" -->');
                    }

                    // 2. Fix mindmap - handle special characters in node labels
                    if (sanitized.startsWith('mindmap')) {
                        const lines = sanitized.split('\n');
                        const processedLines = [];

                        for (let i = 0; i < lines.length; i++) {
                            let line = lines[i];
                            const trimmed = line.trim();

                            // Skip empty lines and the mindmap declaration
                            if (!trimmed || trimmed === 'mindmap') {
                                processedLines.push(line);
                                continue;
                            }

                            // Fix root if missing double parens
                            if (trimmed.startsWith('root') && !line.includes('((')) {
                                line = line.replace(/root\s+(.+)/, 'root(($1))');
                                processedLines.push(line);
                                continue;
                            }

                            // For other nodes, check if they contain special characters that need quoting
                            // Special chars: parentheses, brackets, braces, quotes, colons, semicolons
                            const hasSpecialChars = /[\(\)\[\]\{\}"':;,\s]{2,}/.test(trimmed);
                            const alreadyQuoted = /^["'].*["']$/.test(trimmed) || /^\(.*\)$/.test(trimmed) || /^\[.*\]$/.test(trimmed);

                            if (hasSpecialChars && !alreadyQuoted && trimmed.length > 0) {
                                // Extract indentation and node content
                                const indentMatch = line.match(/^(\s*)/);
                                const indent = indentMatch ? indentMatch[1] : '';
                                const content = trimmed;

                                // Wrap in quotes and preserve the original brackets for styling
                                // Replace inner parentheses that are part of the content with quoted version
                                processedLines.push(indent + '"' + content.replace(/"/g, '\\"') + '"');
                            } else {
                                processedLines.push(line);
                            }
                        }

                        sanitized = processedLines.join('\n');
                    }

                    return sanitized;
                };

                for (let i = 0; i < mermaidBlocks.length; i++) {
                    const block = mermaidBlocks[i];
                    const pre = block.parentElement;
                    const rawCode = block.textContent;
                    const cleanCode = sanitizeMermaid(rawCode);
                    
                    const id = `mermaid-diag-${Date.now()}-${i}`;
                    
                    try {
                        const { svg } = await mermaid.render(id, cleanCode);
                        const container = document.createElement('div');
                        container.className = 'mermaid-diagram';
                        container.innerHTML = svg;
                        pre.parentNode.replaceChild(container, pre);
                    } catch (renderErr) {
                        console.error('Mermaid Render Error:', renderErr);
                        // Final fallback: If rendering failed, try one more time by stripping ALL parentheses from labels
                        try {
                            const lastResort = cleanCode.replace(/\(|\)/g, '');
                            const { svg } = await mermaid.render(`${id}-retry`, lastResort);
                            const container = document.createElement('div');
                            container.className = 'mermaid-diagram';
                            container.innerHTML = svg;
                            pre.parentNode.replaceChild(container, pre);
                        } catch (e) {
                            pre.innerHTML = `<div style="color:red; font-size:12px; padding:10px;">渲染失败: ${renderErr.message}</div>`;
                        }
                    }
                }
            } catch (err) {
                console.error('Mermaid general error:', err);
            }
        }

        // Switch to note tab
        this.switchPanelTab('note');

        // Copy button
        const copyBtn = document.getElementById('btnCopyNote');
        copyBtn.addEventListener('click', async () => {
            try {
                await navigator.clipboard.writeText(note.content);
                const originalHTML = copyBtn.innerHTML;
                copyBtn.innerHTML = `
                    <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2">
                        <polyline points="4,8 6,10 12,4"/>
                    </svg>
                `;
                copyBtn.classList.add('copied');
                setTimeout(() => {
                    copyBtn.innerHTML = originalHTML;
                    copyBtn.classList.remove('copied');
                }, 2000);
                this.setStatus('已复制!');
            } catch (err) {
                this.showError('复制失败');
            }
        });

        // Highlight the selected note in the sidebar
        document.querySelectorAll('.note-item').forEach(el => {
            el.classList.remove('selected');
        });
        const noteItem = document.querySelector(`.note-item[data-id="${note.id}"]`);
        if (noteItem) {
            noteItem.classList.add('selected');
        }
    }

    async deleteNote(id) {
        try {
            await this.api(`/notebooks/${this.currentNotebook.id}/notes/${id}`, {
                method: 'DELETE',
            });
            await this.loadNotes();
            await this.updateCurrentNotebookCounts();

            // If notes_list tab is active or visible, refresh it
            const tabBtnNotesList = document.getElementById('tabBtnNotesList');
            if (tabBtnNotesList && !tabBtnNotesList.classList.contains('hidden')) {
                this.renderNotesCompactGrid();
            }
        } catch (error) {
            this.showError('删除笔记失败');
        }
    }

    // 转换方法
    async handleTransform(type, element) {
        if (!this.currentNotebook) {
            this.showError('请先选择一个笔记本');
            return;
        }

        const sources = await this.api(`/notebooks/${this.currentNotebook.id}/sources`);
        if (sources.length === 0) {
            this.showError('请先添加来源');
            return;
        }

        const customPrompt = document.getElementById('customPrompt').value;
        const typeName = this.noteTypeNameMap[type] || '内容';

        // 1. 开始动画
        if (element) {
            element.classList.add('loading');
        }

        // 2. 添加占位笔记
        const notesContainer = document.getElementById('notesList');
        const template = document.getElementById('noteTemplate');
        const placeholder = template.content.cloneNode(true).querySelector('.note-item');
        
        placeholder.classList.add('placeholder');
        placeholder.querySelector('.note-title').textContent = `正在生成${typeName}...`;
        placeholder.querySelector('.note-preview').textContent = 'AI 正在分析您的来源并撰写笔记，请稍候...';
        placeholder.querySelector('.note-date').textContent = '刚刚';
        placeholder.querySelector('.note-type-badge').textContent = type.toUpperCase();
        
        // 占位符暂不显示删除按钮
        const delBtn = placeholder.querySelector('.btn-delete-note');
        if (delBtn) delBtn.style.display = 'none';
        
        // 如果有“暂无笔记”状态，先清空
        const emptyState = notesContainer.querySelector('.empty-state');
        if (emptyState) emptyState.remove();
        
        notesContainer.prepend(placeholder);
        placeholder.scrollIntoView({ behavior: 'smooth', block: 'nearest' });

        try {
            const sourceIds = sources.map(s => s.id);
            const note = await this.api(`/notebooks/${this.currentNotebook.id}/transform`, {
                method: 'POST',
                body: JSON.stringify({
                    type: type,
                    prompt: customPrompt || undefined,
                    source_ids: sourceIds,
                    length: 'medium',
                    format: 'markdown',
                }),
            });

            // 3. 停止动画并更新占位符
            if (element) element.classList.remove('loading');

            // 替换占位符内容
            placeholder.classList.remove('placeholder');
            placeholder.dataset.id = note.id;
            placeholder.querySelector('.note-title').textContent = note.title;
            
            const plainText = note.content
                .replace(/^#+\s+/gm, '')
                .replace(/\*\*/g, '')
                .replace(/\*/g, '')
                .replace(/`/g, '')
                .replace(/\ \[([^\]]+)\]\([^)]+\)/g, '$1')
                .replace(/\n+/g, ' ')
                .trim();
            
            placeholder.querySelector('.note-preview').textContent = plainText;
            placeholder.querySelector('.note-sources').textContent = `${note.source_ids?.length || 0} 来源`;
            
            // 恢复删除按钮并绑定事件
            if (delBtn) {
                // Apply config to delete button
                if (this.config.allowDelete) {
                    delBtn.style.display = 'flex';
                } else {
                    delBtn.style.display = 'none';
                }
                delBtn.addEventListener('click', (e) => {
                    e.stopPropagation();
                    this.deleteNote(note.id);
                });
            }

            // 绑定查看事件
            placeholder.addEventListener('click', (e) => {
                if (!e.target.closest('.btn-delete-note')) {
                    this.viewNote(note);
                }
            });

            await this.updateCurrentNotebookCounts();
            this.updateFooter();
            document.getElementById('customPrompt').value = '';
            this.setStatus(`成功生成 ${typeName}`);

            // If type is insight, refresh sources list to show the injected insight report
            if (type === 'insight') {
                await this.loadSources();
            }

            // If notes_list tab is active or visible, refresh it
            const tabBtnNotesList = document.getElementById('tabBtnNotesList');
            if (tabBtnNotesList && !tabBtnNotesList.classList.contains('hidden')) {
                this.renderNotesCompactGrid();
            }
        } catch (error) {
            if (element) element.classList.remove('loading');
            placeholder.remove(); // 失败则移除占位符
            this.showError(error.message);
        }
    }

    // 聊天方法
    async loadChatSessions() {
        if (!this.currentNotebook) return;

        try {
            await this.api(`/notebooks/${this.currentNotebook.id}/chat/sessions`);
            const container = document.getElementById('chatMessages');
            container.innerHTML = `
                <div class="chat-welcome">
                    <svg width="40" height="40" viewBox="0 0 40 40" fill="none" stroke="currentColor" stroke-width="1.5">
                        <circle cx="20" cy="12" r="6"/>
                        <path d="M8 38 C8 28 14 22 20 22 C26 22 32 28 32 38"/>
                    </svg>
                    <h3>与来源对话</h3>
                    <p>询问关于笔记本内容的问题</p>
                </div>
            `;
            this.currentChatSession = null;
        } catch (error) {
            console.error('加载对话失败:', error);
        }
    }

    async handleChat(e) {
        e.preventDefault();

        if (!this.currentNotebook) {
            this.showError('请先选择一个笔记本');
            return;
        }

        const input = document.getElementById('chatInput');
        const message = input.value.trim();

        if (!message) return;

        this.addMessage('user', message);
        input.value = '';

        const sources = await this.api(`/notebooks/${this.currentNotebook.id}/sources`);
        if (sources.length === 0) {
            this.addMessage('assistant', '请先为笔记本添加一些来源。');
            return;
        }

        this.setStatus('思考中...');

        const assistantRef = this.addMessage('assistant', '');

        try {
            const streamResult = await this.streamChat(message, assistantRef);
            if (streamResult.used) {
                this.setStatus(streamResult.ok ? '就绪' : '错误');
                return;
            }

            const response = await this.api(`/notebooks/${this.currentNotebook.id}/chat`, {
                method: 'POST',
                body: JSON.stringify({
                    message: message,
                    session_id: this.currentChatSession || undefined,
                }),
            });

            this.updateMessageContent(assistantRef, response.message, true);
            this.updateMessageSources(assistantRef, response.sources);
            this.currentChatSession = response.session_id;
            this.setStatus('就绪');
        } catch (error) {
            this.updateMessageContent(assistantRef, `错误: ${error.message}`, false);
            this.setStatus('错误');
        }
    }

    addMessage(role, content, sources = []) {
        const container = document.getElementById('chatMessages');
        const template = document.getElementById('messageTemplate');

        const welcome = container.querySelector('.chat-welcome');
        if (welcome) welcome.remove();

        const clone = template.content.cloneNode(true);
        const message = clone.querySelector('.chat-message');

        message.dataset.role = role;
        
        const avatar = message.querySelector('.message-avatar');
        avatar.textContent = role === 'assistant' ? 'AI' : '你';

        const messageText = message.querySelector('.message-text');
        const sourcesContainer = message.querySelector('.message-sources');
        const messageRef = {
            role,
            element: message,
            text: messageText,
            sources: sourcesContainer
        };

        this.updateMessageContent(messageRef, content, role === 'assistant');
        this.updateMessageSources(messageRef, sources);

        container.appendChild(clone);
        container.scrollTop = container.scrollHeight;

        return messageRef;
    }

    updateMessageContent(messageRef, content, renderMarkdown = false) {
        if (!messageRef || !messageRef.text) return;
        if (messageRef.role === 'assistant' && renderMarkdown) {
            messageRef.text.innerHTML = marked.parse(content || '');
        } else {
            messageRef.text.textContent = content || '';
        }
    }

    updateMessageSources(messageRef, sources = []) {
        if (!messageRef || !messageRef.sources) return;
        messageRef.sources.innerHTML = '';
        sources.forEach(source => {
            const tag = document.createElement('span');
            tag.className = 'source-tag';
            tag.textContent = source.name || source.id;
            messageRef.sources.appendChild(tag);
        });
    }

    async streamChat(message, assistantRef) {
        if (!window.WebSocket) {
            return { used: false, ok: false };
        }

        const wsUrl = this.getWebSocketURL(`/api/notebooks/${this.currentNotebook.id}/chat/stream`);

        return new Promise((resolve) => {
            let buffer = '';
            let hasDone = false;
            let hasData = false;
            let resolved = false;

            const finish = (used, ok) => {
                if (resolved) return;
                resolved = true;
                resolve({ used, ok });
            };

            const ws = new WebSocket(wsUrl);

            ws.onopen = () => {
                ws.send(JSON.stringify({
                    message: message,
                    session_id: this.currentChatSession || undefined,
                }));
            };

            ws.onmessage = (event) => {
                let data;
                try {
                    data = JSON.parse(event.data);
                } catch (e) {
                    return;
                }

                if (data.type === 'delta') {
                    const chunk = data.content || '';
                    if (chunk) {
                        hasData = true;
                        buffer += chunk;
                        this.updateMessageContent(assistantRef, buffer, false);
                    }
                    return;
                }

                if (data.type === 'done') {
                    hasDone = true;
                    buffer = data.message || buffer;
                    this.updateMessageContent(assistantRef, buffer, true);
                    this.updateMessageSources(assistantRef, data.sources || []);
                    if (data.session_id) {
                        this.currentChatSession = data.session_id;
                    }
                    ws.close();
                    finish(true, true);
                    return;
                }

                if (data.type === 'error') {
                    this.updateMessageContent(assistantRef, `错误: ${data.error || '请求失败'}`, false);
                    ws.close();
                    finish(true, false);
                }
            };

            ws.onerror = () => {
                finish(false, false);
            };

            ws.onclose = () => {
                if (resolved) return;
                if (!hasData) {
                    finish(false, false);
                    return;
                }
                if (!hasDone) {
                    this.updateMessageContent(assistantRef, `${buffer}\n\n(连接中断，请重试)`, false);
                    finish(true, false);
                }
            };
        });
    }

    // UI 方法
    closeModals() {
        document.querySelectorAll('.modal').forEach(m => m.classList.remove('active'));
        document.getElementById('modalOverlay').classList.remove('active');
        this.hideLoading();
    }

    showLoading(text) {
        document.getElementById('loadingText').textContent = text || '处理中...';
        document.getElementById('loadingOverlay').classList.add('active');
    }

    hideLoading() {
        document.getElementById('loadingOverlay').classList.remove('active');
    }

    setStatus(text) {
        document.getElementById('footerStatus').textContent = text;
    }

    showError(message) {
        this.setStatus(`错误: ${message}`);

        const toast = document.createElement('div');
        toast.className = 'error-toast';
        toast.style.cssText = `
            position: fixed; bottom: 60px; right: 20px; padding: 12px 20px;
            background: var(--accent-red); color: white; font-family: var(--font-mono);
            font-size: 0.75rem; border-radius: 4px; box-shadow: var(--shadow-medium);
            animation: slideIn 0.3s ease; z-index: 3000;
        `;
        toast.textContent = message;
        document.body.appendChild(toast);

        setTimeout(() => {
            toast.style.opacity = '0';
            setTimeout(() => toast.remove(), 300);
        }, 3000);
    }

    updateFooter() {
        const sourceCount = document.querySelectorAll('.source-card').length;
        const noteCount = document.querySelectorAll('.note-item').length;
        document.getElementById('footerStats').textContent = `${sourceCount} 来源 · ${noteCount} 笔记`;
    }

    formatDate(dateString) {
        const date = new Date(dateString);
        const now = new Date();
        const diff = now - date;

        if (diff < 60000) return '刚刚';
        if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`;
        if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`;

        return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' });
    }

    async updateCurrentNotebookCounts() {
        if (!this.currentNotebook) return;

        const [sources, notes] = await Promise.all([
            this.api(`/notebooks/${this.currentNotebook.id}/sources`),
            this.api(`/notebooks/${this.currentNotebook.id}/notes`)
        ]);

        const notebookCard = document.querySelector(`.notebook-card[data-id="${this.currentNotebook.id}"]`);
        if (notebookCard) {
            notebookCard.querySelector('.stat-sources').textContent = `${sources.length} 来源`;
            notebookCard.querySelector('.stat-notes').textContent = `${notes.length} 笔记`;
        }
    }
}

// 初始化
document.addEventListener('DOMContentLoaded', () => {
    window.app = new OpenNotebook();
});
