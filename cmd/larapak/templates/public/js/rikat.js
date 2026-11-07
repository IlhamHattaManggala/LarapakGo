/**
 * Rikat UI v1.0 - Reactive Client Component Engine for Larapak & NgapakIn
 * Brings PrimeNG-style declarative bindings and Enterprise Web Components:
 *   - Two-Way Binding: [(model)]="key"
 *   - Property Binding: [property]="expression"
 *   - Event Binding: (event)="expression"
 *   - Rich Components: <l-table>, <l-dialog>, <l-toast>, <l-dropdown>
 * Zero Dependencies - 100% Native Web Standards
 */

(function () {
  'use strict';

  // 1. Reactive State Store
  const State = new Proxy({}, {
    set(target, prop, value) {
      target[prop] = value;
      Rikat.render();
      return true;
    }
  });

  const Rikat = window.Rikat = {
    state: State,
    
    initTooltip() {
      let tip = document.getElementById('rikat-tooltip-bubble');
      if (!tip) {
        tip = document.createElement('div');
        tip.id = 'rikat-tooltip-bubble';
        tip.style.cssText = 'position:fixed;z-index:100000;pointer-events:none;opacity:0;transform:scale(0.95);transition:opacity 0.15s ease, transform 0.15s ease;background:#1c1c21;color:#f4f4f5;font-size:0.75rem;font-weight:500;padding:0.35rem 0.75rem;border-radius:6px;border:1px solid #33333a;box-shadow:0 10px 25px -5px rgba(0,0,0,0.6);white-space:nowrap;display:none;font-family:inherit;';
        if (document.body) document.body.appendChild(tip);
        else document.addEventListener('DOMContentLoaded', () => document.body.appendChild(tip));
      }

      const show = (el, text, pos = 'top') => {
        if (!text) return;
        tip.textContent = text;
        tip.style.display = 'block';
        tip.style.opacity = '0';

        const rect = el.getBoundingClientRect();
        const tipRect = tip.getBoundingClientRect();
        let top = 0;
        let left = 0;

        if (pos === 'bottom') {
          top = rect.bottom + 8;
          left = rect.left + (rect.width / 2) - (tipRect.width / 2);
        } else if (pos === 'left') {
          top = rect.top + (rect.height / 2) - (tipRect.height / 2);
          left = rect.left - tipRect.width - 8;
        } else if (pos === 'right') {
          top = rect.top + (rect.height / 2) - (tipRect.height / 2);
          left = rect.right + 8;
        } else {
          top = rect.top - tipRect.height - 8;
          left = rect.left + (rect.width / 2) - (tipRect.width / 2);
        }

        if (left < 8) left = 8;
        if (left + tipRect.width > window.innerWidth - 8) left = window.innerWidth - tipRect.width - 8;
        if (top < 8) top = rect.bottom + 8;

        tip.style.top = `${top}px`;
        tip.style.left = `${left}px`;
        requestAnimationFrame(() => {
          tip.style.opacity = '1';
          tip.style.transform = 'scale(1)';
        });
      };

      const hide = () => {
        tip.style.opacity = '0';
        tip.style.transform = 'scale(0.95)';
        setTimeout(() => {
          if (tip.style.opacity === '0') tip.style.display = 'none';
        }, 150);
      };

      document.addEventListener('mouseover', (e) => {
        const target = e.target.closest('[pTooltip], [tooltip], [data-tooltip], [ptooltip]');
        if (target) {
          const text = target.getAttribute('pTooltip') || target.getAttribute('tooltip') || target.getAttribute('data-tooltip') || target.getAttribute('ptooltip');
          const pos = target.getAttribute('tooltipPosition') || target.getAttribute('tooltipposition') || 'top';
          show(target, text, pos);
        }
      });

      document.addEventListener('mouseout', (e) => {
        const target = e.target.closest('[pTooltip], [tooltip], [data-tooltip], [ptooltip]');
        if (target) hide();
      });

      document.addEventListener('focusin', (e) => {
        const target = e.target.closest('[pTooltip], [tooltip], [data-tooltip], [ptooltip]');
        if (target) {
          const text = target.getAttribute('pTooltip') || target.getAttribute('tooltip') || target.getAttribute('data-tooltip') || target.getAttribute('ptooltip');
          const pos = target.getAttribute('tooltipPosition') || target.getAttribute('tooltipposition') || 'top';
          show(target, text, pos);
        }
      });

      document.addEventListener('focusout', (e) => {
        const target = e.target.closest('[pTooltip], [tooltip], [data-tooltip], [ptooltip]');
        if (target) hide();
      });
    },
    init() {
      this.bindInputs();
      this.initTooltip();
            this.bindEvents();
      this.render();
    },

    set(key, val) {
      State[key] = val;
    },

    get(key) {
      return State[key];
    },

    // Evaluates expression in context of Rikat.state
    eval(expr, context = {}) {
      try {
        const fn = new Function('state', 'ctx', `with(state) { with(ctx) { return (${expr}); } }`);
        return fn(State, context);
      } catch (e) {
        return undefined;
      }
    },

    // Runs action expression in context of Rikat.state
    run(action, context = {}) {
      try {
        const fn = new Function('state', 'ctx', `with(state) { with(ctx) { ${action}; } }`);
        fn(State, context);
        this.render();
      } catch (e) {
        console.error('Rikat Action Error:', e);
      }
    },

    bindInputs() {
      // Find all elements with [(model)]
      document.querySelectorAll('[\\(\\[model\\]\\)]').forEach(el => {
        const key = el.getAttribute('([(model)])');
        if (!key) return;

        // Initialize state if empty
        if (State[key] === undefined) {
          if (el.type === 'checkbox') State[key] = el.checked;
          else State[key] = el.value || '';
        }

        // Listen for input changes
        const eventName = (el.type === 'checkbox' || el.tagName === 'SELECT') ? 'change' : 'input';
        el.addEventListener(eventName, () => {
          if (el.type === 'checkbox') State[key] = el.checked;
          else State[key] = el.value;
        });
      });
    },

    bindEvents() {
      // Walk attributes looking for (event)
      const allElements = document.querySelectorAll('*');
      allElements.forEach(el => {
        Array.from(el.attributes).forEach(attr => {
          const match = attr.name.match(/^\(([\w-]+)\)$/);
          if (match) {
            const eventName = match[1];
            const action = attr.value;
            el.addEventListener(eventName, (e) => {
              Rikat.run(action, { $event: e });
            });
          }
        });
      });
    },

    render() {
      // 1. Update [(model)] values
      document.querySelectorAll('[\\(\\[model\\]\\)]').forEach(el => {
        const key = el.getAttribute('([(model)])');
        const val = State[key];
        if (val !== undefined) {
          if (el.type === 'checkbox') el.checked = Boolean(val);
          else if (el.value !== val) el.value = val;
        }
      });

      // 2. Update [property] bindings
      document.querySelectorAll('*').forEach(el => {
        Array.from(el.attributes).forEach(attr => {
          const match = attr.name.match(/^\[([\w-]+)\]$/);
          if (match) {
            const prop = match[1];
            const expr = attr.value;
            const res = Rikat.eval(expr);

            if (prop === 'disabled') {
              el.disabled = Boolean(res);
            } else if (prop === 'visible' || prop === 'show') {
              el.style.display = res ? '' : 'none';
            } else if (prop === 'class') {
              if (typeof res === 'string') el.className = res;
            } else if (prop === 'text' || prop === 'html') {
              el.innerText = res ?? '';
            } else {
              el[prop] = res;
            }
          }
        });
      });

      // 3. Update text interpolation {{ key }}
      if (!this._templateCache) {
        this._templateCache = new Map();
        document.querySelectorAll('*').forEach(el => {
          if (el.children.length === 0 && el.textContent.includes('{{')) {
            this._templateCache.set(el, el.textContent);
          }
        });
      }

      this._templateCache.forEach((origText, el) => {
        const newText = origText.replace(/\{\{\s*([a-zA-Z0-9_$.]+)\s*\}\}/g, (_, expr) => {
          const val = Rikat.eval(expr);
          return val !== undefined ? val : '';
        });
        if (el.textContent !== newText) {
          el.textContent = newText;
        }
      });

      // Notify custom elements
      document.querySelectorAll('l-table, l-dialog, r-table, r-dialog, l-drawer, r-drawer').forEach(el => {
        if (typeof el.onStateUpdate === 'function') {
          el.onStateUpdate();
        }
      });
    },

    toast(title, message, type = 'info') {
      let container = document.getElementById('rikat-toast-container');
      if (!container) {
        container = document.createElement('div');
        container.id = 'rikat-toast-container';
        container.style.cssText = 'position:fixed;top:1.5rem;right:1.5rem;z-index:9999;display:flex;flex-direction:column;gap:0.75rem;max-width:360px;';
        document.body.appendChild(container);
      }

      const toast = document.createElement('div');
      toast.className = `alert alert-${type} shadow-lg`;
      toast.style.cssText = 'margin:0;animation:rikatSlideIn 0.25s ease;';
      toast.innerHTML = `
        <div style="flex-grow:1;">
          <div style="font-weight:600;margin-bottom:2px;">${title}</div>
          <div style="font-size:0.8rem;opacity:0.9;">${message}</div>
        </div>
      `;
      container.appendChild(toast);

      setTimeout(() => {
        toast.style.opacity = '0';
        toast.style.transform = 'translateX(20px)';
        toast.style.transition = 'all 0.25s ease';
        setTimeout(() => toast.remove(), 250);
      }, 3500);
    }
  };

  // Add Toast Animation Keyframes & Custom Elements Base CSS
  const style = document.createElement('style');
  style.textContent = `
    
    @keyframes rikatSpin {
      from { transform: rotate(0deg); }
      to { transform: rotate(360deg); }
    }
    .c-btn-rounded { border-radius: 9999px !important; }
    .c-btn-lg { padding: 0.75rem 1.5rem; font-size: 1rem; }
    .c-btn-icon-only { width: 38px; height: 38px; padding: 0; display: inline-flex; align-items: center; justify-content: center; }
    .c-badge-pill { border-radius: 9999px !important; }
    .c-badge-lg { padding: 0.35rem 0.75rem; font-size: 0.85rem; }
    .c-kbd { background: #212126; border: 1px solid #33333d; border-bottom: 2px solid #3f3f4a; border-radius: 5px; padding: 2px 6px; font-size: 0.75rem; font-weight: 600; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; color: #f4f4f5; box-shadow: 0 1px 2px rgba(0,0,0,0.4); }

    
    @keyframes rikatSpin {
      from { transform: rotate(0deg); }
      to { transform: rotate(360deg); }
    }
    .c-btn-rounded { border-radius: 9999px !important; }
    .c-btn-lg { padding: 0.75rem 1.5rem; font-size: 1rem; }
    .c-btn-icon-only { width: 38px; height: 38px; padding: 0; display: inline-flex; align-items: center; justify-content: center; }
    .c-badge-pill { border-radius: 9999px !important; }
    .c-badge-lg { padding: 0.35rem 0.75rem; font-size: 0.85rem; }
    .c-kbd { background: #212126; border: 1px solid #33333d; border-bottom: 2px solid #3f3f4a; border-radius: 5px; padding: 2px 6px; font-size: 0.75rem; font-weight: 600; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; color: #f4f4f5; box-shadow: 0 1px 2px rgba(0,0,0,0.4); }

    
    @keyframes rikatShimmer {
      0% { background-position: 200% 0; }
      100% { background-position: -200% 0; }
    }

    @keyframes rikatSlideIn {
      from { transform: translateX(40px); opacity: 0; }
      to { transform: translateX(0); opacity: 1; }
    }
    l-icon, r-icon {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      vertical-align: middle;
      line-height: 0;
    }
    l-button, r-button {
      display: inline-block;
    }
    .c-btn {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      gap: 0.5rem;
      padding: 0.55rem 1.15rem;
      border-radius: 6px;
      font-size: 0.875rem;
      font-weight: 500;
      cursor: pointer;
      transition: all 0.15s ease;
      border: 1px solid transparent;
      font-family: inherit;
      text-decoration: none;
      line-height: 1.4;
      box-sizing: border-box;
    }
    .c-btn-sm { padding: 0.35rem 0.75rem; font-size: 0.8rem; }
    .c-btn-primary { background: #ff2d20; color: #fff; border-color: #ff2d20; }
    .c-btn-primary:hover { background: #e02418; border-color: #e02418; }
    .c-btn-secondary { background: #27272a; color: #f4f4f5; border-color: #3f3f46; }
    .c-btn-secondary:hover { background: #3f3f46; border-color: #52525b; }
    .c-btn-outline { background: transparent; color: #f4f4f5; border-color: #3f3f46; }
    .c-btn-outline:hover { background: #27272a; border-color: #52525b; }
    .c-btn-danger { background: #dc2626; color: #fff; border-color: #dc2626; }
    .c-btn-danger:hover { background: #b91c1c; border-color: #b91c1c; }
    .c-btn-ghost { background: transparent; color: #a1a1aa; border-color: transparent; }
    .c-btn-ghost:hover { background: rgba(255,255,255,0.06); color: #fff; }
    .c-btn:disabled, .c-btn[disabled] { opacity: 0.5; cursor: not-allowed; pointer-events: none; }

    /* Accordion */
    l-accordion, r-accordion { display: block; }
    .c-accordion-item { border: 1px solid #27272a; border-radius: 8px; margin-bottom: 0.5rem; background: #161619; overflow: hidden; }
    .c-accordion-header { width: 100%; display: flex; align-items: center; justify-content: space-between; padding: 0.85rem 1.25rem; font-size: 0.9rem; font-weight: 600; color: #fff; cursor: pointer; background: transparent; border: none; text-align: left; font-family: inherit; }
    .c-accordion-header:hover { background: rgba(255,255,255,0.03); }
    .c-accordion-body { padding: 0.85rem 1.25rem; font-size: 0.85rem; color: #a1a1aa; border-top: 1px solid #27272a; line-height: 1.6; }

    /* Tabs */
    l-tabs, r-tabs { display: block; }
    .c-tabs { display: flex; border-bottom: 1px solid #27272a; gap: 0.25rem; margin-bottom: 1.25rem; }
    .c-tab { all: unset; padding: 0.65rem 1.15rem; font-size: 0.85rem; font-weight: 500; color: #a1a1aa; cursor: pointer; border-bottom: 2px solid transparent; transition: all 0.15s ease; border-radius: 4px 4px 0 0; }
    .c-tab:hover { color: #fff; background: rgba(255,255,255,0.03); }
    .c-tab.active { color: #ff2d20; border-bottom-color: #ff2d20; font-weight: 600; }

    /* Stepper */
    l-stepper, r-stepper { display: block; }
    .c-stepper { display: flex; align-items: center; width: 100%; gap: 0.5rem; }
    .c-step-item { display: flex; align-items: center; gap: 0.5rem; flex-shrink: 0; }
    .c-step-bubble { width: 30px; height: 30px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 0.75rem; font-weight: 700; flex-shrink: 0; }
    .c-step-bubble.completed { background: #10b981; color: #fff; }
    .c-step-bubble.active { background: #ff2d20; color: #fff; box-shadow: 0 0 0 3px rgba(255,45,32,0.25); }
    .c-step-bubble.pending { background: #27272a; color: #71717a; }
    .c-step-label { font-size: 0.8rem; font-weight: 600; color: #fff; white-space: nowrap; }
    .c-step-label.text-dim { color: #71717a; }
    .c-step-line { flex: 1; height: 2px; background: #27272a; margin: 0 0.5rem; min-width: 20px; }
    .c-step-line.completed { background: #10b981; }

    /* Timeline */
    l-timeline, r-timeline { display: block; }
    .c-timeline { position: relative; padding-left: 2rem; border-left: 2px solid #27272a; margin-left: 0.75rem; }
    .c-timeline-item { position: relative; margin-bottom: 1.75rem; }
    .c-timeline-item:last-child { margin-bottom: 0; }
    .c-timeline-dot { position: absolute; left: calc(-2rem - 7px); top: 3px; width: 12px; height: 12px; border-radius: 50%; background: #ff2d20; border: 2px solid #121214; }
    .c-timeline-time { font-size: 0.75rem; color: #71717a; margin-bottom: 0.25rem; }
    .c-timeline-title { font-size: 0.9rem; font-weight: 600; color: #fff; margin-bottom: 0.25rem; }
    .c-timeline-desc { font-size: 0.825rem; color: #a1a1aa; line-height: 1.5; }

    /* Card */
    l-card, r-card { display: block; }
    .c-card { background: #161619; border: 1px solid #27272a; border-radius: 10px; padding: 1.5rem; }

  `;
  if (document.head) {
    document.head.appendChild(style);
  } else {
    document.addEventListener('DOMContentLoaded', () => document.head.appendChild(style));
  }

  // ==========================================================================
  // Web Component 1: <l-dialog> (PrimeNG-style modal)
  // Usage: <l-dialog [(visible)]="showModal" header="Judul">...</l-dialog>
  // ==========================================================================

  // Safe Custom Element Registration helper (handles duplicate constructor definitions cleanly)
  function defElement(tag, cls) {
    if (typeof customElements === 'undefined' || customElements.get(tag)) return;
    try {
      customElements.define(tag, cls);
    } catch (_) {
      try {
        customElements.define(tag, class extends cls {});
      } catch (__) {}
    }
  }

  class LDialog extends HTMLElement {
    connectedCallback() {
      this.visibleKey = this.getAttribute('[(visible)]') || this.getAttribute('[visible]') || this.getAttribute('visible') || this.getAttribute('([(visible)])') || null;
      this.headerText = this.getAttribute('header') || 'Dialog';
      this.style.display = 'contents';
      
      // Delay initialization until child elements are fully parsed by HTML streaming parser
      if (document.readyState === 'loading') {
        const handler = () => {
          document.removeEventListener('DOMContentLoaded', handler);
          this.initDialog();
        };
        document.addEventListener('DOMContentLoaded', handler);
      } else {
        setTimeout(() => this.initDialog(), 0);
      }
    }

    initDialog() {
      if (this._initialized) return;
      this._initialized = true;

      // Extract all children to preserve their live elements and event listeners
      const frag = document.createDocumentFragment();
      while (this.firstChild) {
        frag.appendChild(this.firstChild);
      }

      this.innerHTML = `
        <div class="rikat-modal-backdrop" style="display:none;position:fixed;inset:0;background:rgba(0,0,0,0.75);backdrop-filter:blur(6px);z-index:9999;align-items:center;justify-content:center;padding:1.5rem;box-sizing:border-box;">
          <div class="rikat-modal-card" style="width:100%;max-width:520px;background:#18181b;border:1px solid #27272a;border-radius:12px;box-shadow:0 25px 50px -12px rgba(0,0,0,0.7);animation:rikatSlideIn 0.2s cubic-bezier(0.16,1,0.3,1);overflow:hidden;box-sizing:border-box;">
            <div class="rikat-modal-header" style="display:flex;align-items:center;justify-content:space-between;border-bottom:1px solid #27272a;padding:1rem 1.25rem;background:#141417;">
              <span class="rikat-modal-title" style="font-weight:600;font-size:1.05rem;color:#f4f4f5;letter-spacing:-0.01em;">${this.headerText}</span>
              <button type="button" class="rikat-dialog-close" style="all:unset;cursor:pointer;color:#a1a1aa;display:flex;align-items:center;justify-content:center;width:28px;height:28px;border-radius:6px;font-size:1.25rem;line-height:1;transition:all 0.15s ease;" onmouseover="this.style.color='#fff';this.style.background='rgba(255,255,255,0.08)'" onmouseout="this.style.color='#a1a1aa';this.style.background='transparent'">&times;</button>
            </div>
            <div class="rikat-modal-body" style="padding:1.5rem;color:#d4d4d8;box-sizing:border-box;">
            </div>
          </div>
        </div>
      `;

      // Append preserved children into the modal body
      this.querySelector('.rikat-modal-body').appendChild(frag);

      // Close button handler
      this.querySelector('.rikat-dialog-close').addEventListener('click', () => {
        this.close();
      });

      // Close on backdrop click (outside card)
      const backdrop = this.querySelector('.rikat-modal-backdrop');
      backdrop.addEventListener('click', (e) => {
        if (e.target === backdrop) {
          this.close();
        }
      });

      // Escape key handler
      document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape' && this.isOpen()) {
          this.close();
        }
      });

      this.onStateUpdate();
    }

    isOpen() {
      const modal = this.querySelector('.rikat-modal-backdrop');
      return modal && modal.style.display === 'flex';
    }

    close() {
      if (this.visibleKey) {
        Rikat.set(this.visibleKey, false);
      } else {
        const modal = this.querySelector('.rikat-modal-backdrop');
        if (modal) modal.style.display = 'none';
        document.body.style.overflow = '';
      }
    }

    onStateUpdate() {
      if (!this.visibleKey) return;
      const isVisible = Boolean(Rikat.get(this.visibleKey));
      const modal = this.querySelector('.rikat-modal-backdrop');
      if (modal) {
        modal.style.display = isVisible ? 'flex' : 'none';
        if (isVisible) {
          document.body.style.overflow = 'hidden';
        } else {
          document.body.style.overflow = '';
        }
      }
    }
  }
  defElement('l-dialog', LDialog);
  defElement('r-dialog', LDialog);

  // ==========================================================================
  // Web Component 2: <l-table> (PrimeNG-style enterprise DataTable)
  // Usage: <l-table [data]="users" [paginator]="bener" [rows]="5"></l-table>
  // ==========================================================================
  class LTable extends HTMLElement {
    connectedCallback() {
      this.dataKey = this.getAttribute('[data]') || this.getAttribute('[value]');
      this.hasPaginator = this.getAttribute('[paginator]') === 'true' || this.getAttribute('[paginator]') === 'bener' || this.hasAttribute('paginator');
      this.rowsPerPage = parseInt(this.getAttribute('[rows]') || '5', 10);
      this.currentPage = 1;
      this.sortField = null;
      this.sortAsc = true;
      this.searchQuery = '';
      this.renderTable();
    }

    onStateUpdate() {
      this.updateRows();
    }

    getData() {
      let data = Rikat.eval(this.dataKey) || [];
      if (!Array.isArray(data)) return [];

      // Filter
      if (this.searchQuery) {
        const q = this.searchQuery.toLowerCase();
        data = data.filter(item => {
          return Object.values(item).some(val => String(val).toLowerCase().includes(q));
        });
      }

      // Sort
      if (this.sortField) {
        data = [...data].sort((a, b) => {
          const valA = a[this.sortField];
          const valB = b[this.sortField];
          if (valA < valB) return this.sortAsc ? -1 : 1;
          if (valA > valB) return this.sortAsc ? 1 : -1;
          return 0;
        });
      }

      return data;
    }

    renderTable() {
      this.innerHTML = `
        <div class="card border-subtle" style="overflow:hidden;">
          <div class="card-header flex items-center justify-between" style="background:var(--bg-surface-elevated);gap:1rem;">
            <input type="text" class="input rikat-table-search" placeholder="Goleki data..." style="max-width:260px;font-size:0.8rem;padding:0.4rem 0.75rem;">
            <span class="rikat-table-count text-xs text-muted"></span>
          </div>
          <div style="overflow-x:auto;">
            <table class="table table-hover w-full">
              <thead class="rikat-thead"></thead>
              <tbody class="rikat-tbody"></tbody>
            </table>
          </div>
          <div class="rikat-table-pagination card-footer flex items-center justify-between" style="display:${this.hasPaginator ? 'flex' : 'none'};">
            <span class="text-xs text-muted rikat-page-info"></span>
            <div class="flex gap-2">
              <button type="button" class="btn btn-secondary btn-sm rikat-btn-prev">&larr; Sedurunge</button>
              <button type="button" class="btn btn-secondary btn-sm rikat-btn-next">Sabanjure &rarr;</button>
            </div>
          </div>
        </div>
      `;

      this.querySelector('.rikat-table-search').addEventListener('input', (e) => {
        this.searchQuery = e.target.value;
        this.currentPage = 1;
        this.updateRows();
      });

      this.querySelector('.rikat-btn-prev').addEventListener('click', () => {
        if (this.currentPage > 1) {
          this.currentPage--;
          this.updateRows();
        }
      });

      this.querySelector('.rikat-btn-next').addEventListener('click', () => {
        const totalPages = Math.ceil(this.getData().length / this.rowsPerPage);
        if (this.currentPage < totalPages) {
          this.currentPage++;
          this.updateRows();
        }
      });

      this.updateRows();
    }

    updateRows() {
      const allData = this.getData();
      const tbody = this.querySelector('.rikat-tbody');
      const thead = this.querySelector('.rikat-thead');
      const countLabel = this.querySelector('.rikat-table-count');
      const pageInfo = this.querySelector('.rikat-page-info');

      if (!tbody || !thead) return;

      countLabel.textContent = `Total: ${allData.length} data`;

      if (allData.length === 0) {
        thead.innerHTML = '';
        tbody.innerHTML = '<tr><td colspan="100%" class="text-center text-muted" style="padding:2rem;">Ora ana data sing ditemokake</td></tr>';
        if (pageInfo) pageInfo.textContent = 'Kaca 0 saka 0';
        return;
      }

      // Auto-extract columns from first object
      const columns = Object.keys(allData[0]);

      // Render Header
      thead.innerHTML = `
        <tr>
          ${columns.map(col => `
            <th class="cursor-pointer select-none" data-col="${col}">
              ${col.toUpperCase()} ${this.sortField === col ? (this.sortAsc ? '▲' : '▼') : ''}
            </th>
          `).join('')}
        </tr>
      `;

      // Header click for sorting
      thead.querySelectorAll('th').forEach(th => {
        th.addEventListener('click', () => {
          const col = th.getAttribute('data-col');
          if (this.sortField === col) {
            this.sortAsc = !this.sortAsc;
          } else {
            this.sortField = col;
            this.sortAsc = true;
          }
          this.updateRows();
        });
      });

      // Pagination slice
      const start = (this.currentPage - 1) * this.rowsPerPage;
      const end = this.hasPaginator ? start + this.rowsPerPage : allData.length;
      const pageData = allData.slice(start, end);

      const totalPages = Math.ceil(allData.length / this.rowsPerPage) || 1;
      if (pageInfo) pageInfo.textContent = `Kaca ${this.currentPage} saka ${totalPages}`;

      // Render Rows
      tbody.innerHTML = pageData.map(row => `
        <tr class="cursor-pointer">
          ${columns.map(col => `<td>${row[col] !== undefined ? row[col] : ''}</td>`).join('')}
        </tr>
      `).join('');

      // Row click listener
      tbody.querySelectorAll('tr').forEach((tr, idx) => {
        tr.addEventListener('click', () => {
          const item = pageData[idx];
          const onSelectAttr = this.getAttribute('(onSelect)');
          if (onSelectAttr) {
            Rikat.run(onSelectAttr, { $event: item });
          }
        });
      });
    }
  }
  defElement('l-table', LTable);
  defElement('r-table', LTable);

  // 60+ Modern Feather/Lucide style SVG paths for Larapak & Rikat
  Rikat.icons = {
    'search': '<circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line>',
    'copy': '<rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>',
    'check': '<polyline points="20 6 9 17 4 12"></polyline>',
    'check-circle': '<path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline>',
    'x': '<line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line>',
    'x-circle': '<circle cx="12" cy="12" r="10"></circle><line x1="15" y1="9" x2="9" y2="15"></line><line x1="9" y1="9" x2="15" y2="15"></line>',
    'arrow-right': '<line x1="5" y1="12" x2="19" y2="12"></line><polyline points="12 5 19 12 12 19"></polyline>',
    'arrow-left': '<line x1="19" y1="12" x2="5" y2="12"></line><polyline points="12 19 5 12 12 5"></polyline>',
    'arrow-up': '<line x1="12" y1="19" x2="12" y2="5"></line><polyline points="5 12 12 5 19 12"></polyline>',
    'arrow-down': '<line x1="12" y1="5" x2="12" y2="19"></line><polyline points="19 12 12 19 5 12"></polyline>',
    'chevron-down': '<polyline points="6 9 12 15 18 9"></polyline>',
    'chevron-up': '<polyline points="18 15 12 9 6 15"></polyline>',
    'chevron-right': '<polyline points="9 18 15 12 9 6"></polyline>',
    'chevron-left': '<polyline points="15 18 9 12 15 6"></polyline>',
    'user': '<path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle>',
    'users': '<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle><path d="M23 21v-2a4 4 0 0 0-3-3.87"></path><path d="M16 3.13a4 4 0 0 1 0 7.75"></path>',
    'user-plus': '<path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle><line x1="20" y1="8" x2="20" y2="14"></line><line x1="23" y1="11" x2="17" y2="11"></line>',
    'trash': '<polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>',
    'bell': '<path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"></path><path d="M13.73 21a2 2 0 0 1-3.46 0"></path>',
    'settings': '<circle cx="12" cy="12" r="3"></circle><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path>',
    'plus': '<line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line>',
    'minus': '<line x1="5" y1="12" x2="19" y2="12"></line>',
    'download': '<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line>',
    'upload': '<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="17 8 12 3 7 8"></polyline><line x1="12" y1="3" x2="12" y2="15"></line>',
    'external-link': '<path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"></path><polyline points="15 3 21 3 21 9"></polyline><line x1="10" y1="14" x2="21" y2="3"></line>',
    'eye': '<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path><circle cx="12" cy="12" r="3"></circle>',
    'eye-off': '<path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"></path><line x1="1" y1="1" x2="23" y2="23"></line>',
    'lock': '<rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect><path d="M7 11V7a5 5 0 0 1 10 0v4"></path>',
    'unlock': '<rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect><path d="M7 11V7a5 5 0 0 1 9.9-1"></path>',
    'mail': '<path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"></path><polyline points="22,6 12,13 2,6"></polyline>',
    'phone': '<path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z"></path>',
    'calendar': '<rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect><line x1="16" y1="2" x2="16" y2="6"></line><line x1="8" y1="2" x2="8" y2="6"></line><line x1="3" y1="10" x2="21" y2="10"></line>',
    'filter': '<polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3"></polygon>',
    'refresh': '<polyline points="23 4 23 10 17 10"></polyline><polyline points="1 20 1 14 7 14"></polyline><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path>',
    'star': '<polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"></polygon>',
    'heart': '<path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"></path>',
    'shield': '<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path>',
    'alert-triangle': '<path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path><line x1="12" y1="9" x2="12" y2="13"></line><line x1="12" y1="17" x2="12.01" y2="17"></line>',
    'info': '<circle cx="12" cy="12" r="10"></circle><line x1="12" y1="16" x2="12" y2="12"></line><line x1="12" y1="8" x2="12.01" y2="8"></line>',
    'help-circle': '<circle cx="12" cy="12" r="10"></circle><path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"></path><line x1="12" y1="17" x2="12.01" y2="17"></line>',
    'folder': '<path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path>',
    'file': '<path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"></path><polyline points="13 2 13 9 20 9"></polyline>',
      'file-text': '<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="16" y1="13" x2="8" y2="13"></line><line x1="16" y1="17" x2="8" y2="17"></line>',
      'file-code': '<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><polyline points="10 13 8 15 10 17"></polyline><polyline points="14 13 16 15 14 17"></polyline>',
    'code': '<polyline points="16 18 22 12 16 6"></polyline><polyline points="8 6 2 12 8 18"></polyline>',
    'terminal': '<polyline points="4 17 10 11 4 5"></polyline><line x1="12" y1="19" x2="20" y2="19"></line>',
    'database': '<ellipse cx="12" cy="5" rx="9" ry="3"></ellipse><path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"></path><path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"></path>',
    'server': '<rect x="2" y="2" width="20" height="8" rx="2" ry="2"></rect><rect x="2" y="14" width="20" height="8" rx="2" ry="2"></rect><line x1="6" y1="6" x2="6.01" y2="6"></line><line x1="6" y1="18" x2="6.01" y2="18"></line>',
    'zap': '<polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon>',
    'clock': '<circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline>',
    'tag': '<path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"></path><line x1="7" y1="7" x2="7.01" y2="7"></line>',
    'bookmark': '<path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z"></path>',
      'refresh-cw': '<polyline points="23 4 23 10 17 10"></polyline><polyline points="1 20 1 14 7 14"></polyline><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path>',
      'edit': '<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>',
      'check-square': '<polyline points="9 11 12 14 22 4"></polyline><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"></path>',
      'columns': '<path d="M12 3h7a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-7m0-18H5a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h7m0-18v18"></path>',
      'droplet': '<path d="M12 2.69l5.66 5.66a8 8 0 1 1-11.31 0z"></path>',
      'maximize': '<path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3"></path>',
      'sidebar': '<rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect><line x1="9" y1="3" x2="9" y2="21"></line>',
      'more-horizontal': '<circle cx="12" cy="12" r="1"></circle><circle cx="19" cy="12" r="1"></circle><circle cx="5" cy="12" r="1"></circle>',
      'award': '<circle cx="12" cy="8" r="7"></circle><polyline points="8.21 13.89 7 23 12 20 17 23 15.79 13.88"></polyline>',
      'alert-circle': '<circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line>',
      'loader': '<line x1="12" y1="2" x2="12" y2="6"></line><line x1="12" y1="18" x2="12" y2="22"></line><line x1="4.93" y1="4.93" x2="7.76" y2="7.76"></line><line x1="16.24" y1="16.24" x2="19.07" y2="19.07"></line><line x1="2" y1="12" x2="6" y2="12"></line><line x1="18" y1="12" x2="22" y2="12"></line><line x1="4.93" y1="19.07" x2="7.76" y2="16.24"></line><line x1="16.24" y1="7.76" x2="19.07" y2="4.93"></line>',
      'trending-up': '<polyline points="23 6 13.5 15.5 8.5 10.5 1 18"></polyline><polyline points="17 6 23 6 23 12"></polyline>',
      'command': '<path d="M18 3a3 3 0 0 0-3 3v12a3 3 0 0 0 3 3 3 3 0 0 0 3-3 3 3 0 0 0-3-3H6a3 3 0 0 0-3 3 3 3 0 0 0 3 3 3 3 0 0 0 3-3V6a3 3 0 0 0-3-3 3 3 0 0 0-3 3 3 3 0 0 0 3 3h12a3 3 0 0 0 3-3 3 3 0 0 0-3-3z"></path>',
      'square': '<rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>',
      'share-2': '<circle cx="18" cy="5" r="3"></circle><circle cx="6" cy="12" r="3"></circle><circle cx="18" cy="19" r="3"></circle><line x1="8.59" y1="13.51" x2="15.42" y2="17.49"></line><line x1="15.41" y1="6.51" x2="8.59" y2="10.49"></line>',
      'book-open': '<path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"></path><path d="M22 3h-6a4 4 0 0 1-4 4v14a3 3 0 0 1 3-3h7z"></path>',
      'book': '<path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"></path>',
    'sun': '<circle cx="12" cy="12" r="5"></circle><line x1="12" y1="1" x2="12" y2="3"></line><line x1="12" y1="21" x2="12" y2="23"></line><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line><line x1="1" y1="12" x2="3" y2="12"></line><line x1="21" y1="12" x2="23" y2="12"></line><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line>',
    'moon': '<path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path>',
    'github': '<path fill-rule="evenodd" clip-rule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.53 1.032 1.53 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"></path>',
    'share': '<circle cx="18" cy="5" r="3"></circle><circle cx="6" cy="12" r="3"></circle><circle cx="18" cy="19" r="3"></circle><line x1="8.59" y1="13.51" x2="15.42" y2="17.49"></line><line x1="15.41" y1="6.51" x2="8.59" y2="10.49"></line>',
    'link': '<path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"></path><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"></path>',
    'layout': '<rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect><line x1="3" y1="9" x2="21" y2="9"></line><line x1="9" y1="21" x2="9" y2="9"></line>',
    'grid': '<rect x="3" y="3" width="7" height="7"></rect><rect x="14" y="3" width="7" height="7"></rect><rect x="14" y="14" width="7" height="7"></rect><rect x="3" y="14" width="7" height="7"></rect>',
    'list': '<line x1="8" y1="6" x2="21" y2="6"></line><line x1="8" y1="12" x2="21" y2="12"></line><line x1="8" y1="18" x2="21" y2="18"></line><line x1="3" y1="6" x2="3.01" y2="6"></line><line x1="3" y1="12" x2="3.01" y2="12"></line><line x1="3" y1="18" x2="3.01" y2="18"></line>',
    'sliders': '<line x1="4" y1="21" x2="4" y2="14"></line><line x1="4" y1="10" x2="4" y2="3"></line><line x1="12" y1="21" x2="12" y2="12"></line><line x1="12" y1="8" x2="12" y2="3"></line><line x1="20" y1="21" x2="20" y2="16"></line><line x1="20" y1="12" x2="20" y2="3"></line><line x1="1" y1="14" x2="7" y2="14"></line><line x1="9" y1="8" x2="15" y2="8"></line><line x1="17" y1="16" x2="23" y2="16"></line>',
    'inbox': '<polyline points="22 12 16 12 14 15 10 15 8 12 2 12"></polyline><path d="M5.45 5.11L2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11z"></path>',
    'message-square': '<path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>',
    'send': '<line x1="22" y1="2" x2="11" y2="13"></line><polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>',
    'activity': '<polyline points="22 12 18 12 15 21 9 3 6 12 2 12"></polyline>',
    'layers': '<polygon points="12 2 2 7 12 12 22 7 12 2"></polygon><polyline points="2 17 12 22 22 17"></polyline><polyline points="2 12 12 17 22 12"></polyline>',
    'pie-chart': '<path d="M21.21 15.89A10 10 0 1 1 8 2.83"></path><path d="M22 12A10 10 0 0 0 12 2v10z"></path>',
    'bar-chart-2': '<line x1="18" y1="20" x2="18" y2="10"></line><line x1="12" y1="20" x2="12" y2="4"></line><line x1="6" y1="20" x2="6" y2="14"></line>',
    'box': '<path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"></path><polyline points="3.27 6.96 12 12.01 20.73 6.96"></polyline><line x1="12" y1="22.08" x2="12" y2="12"></line>'
  };

  // ==========================================================================
  // Web Component 3: <l-icon> & <r-icon>
  // Usage: <l-icon name="search" size="16"></l-icon>
  // ==========================================================================
  class LIcon extends HTMLElement {
    static get observedAttributes() { return ['name', 'size', 'color', 'stroke']; }
    connectedCallback() { this.render(); }
    attributeChangedCallback() { this.render(); }
    render() {
      const name = this.getAttribute('name') || '';
      const size = this.getAttribute('size') || '16';
      const color = this.getAttribute('color') || 'currentColor';
      const stroke = this.getAttribute('stroke') || '2';
      const inner = (Rikat.icons && Rikat.icons[name]) ? Rikat.icons[name] : '<circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line>';
      this.innerHTML = `<svg width="${size}" height="${size}" viewBox="0 0 24 24" fill="none" stroke="${color}" stroke-width="${stroke}" stroke-linecap="round" stroke-linejoin="round" style="width:${size}px;height:${size}px;flex-shrink:0;display:block;">${inner}</svg>`;
    }
  }
  defElement('l-icon', LIcon);
  defElement('r-icon', LIcon);

  // ==========================================================================
  // Rikat UI Enterprise Web Components (PrimeNG Style)
  // Declarative custom tags: <l-button>, <l-card>, <l-badge>, <l-alert>,
  // <l-input>, <l-switch>, <l-progress>, <l-avatar>, <l-skeleton>, <l-stat>,
  // <l-empty-state>, <l-accordion>, <l-tabs>, <l-breadcrumb>, <l-stepper>, etc.
  // ==========================================================================


  class LTooltip extends HTMLElement {
    connectedCallback() {
      this.style.display = 'inline-block';
      const text = this.getAttribute('text') || this.getAttribute('content') || '';
      const pos = this.getAttribute('position') || 'top';
      this.setAttribute('pTooltip', text);
      this.setAttribute('tooltipPosition', pos);
    }
  }
  defElement('l-tooltip', LTooltip);
  defElement('r-tooltip', LTooltip);

  // ==========================================================================
  // Rikat UI Enterprise Web Components (PrimeNG Style)
  // ==========================================================================

  // Robust onReady helper: guarantees DOM and child nodes are fully parsed
  function onReady(el, fn) {
    const run = () => {
      if (!el.isConnected) return;
      fn.call(el);
    };

    if (document.readyState === 'loading') {
      document.addEventListener('DOMContentLoaded', run, { once: true });
    } else {
      setTimeout(run, 0);
    }
  }

  // 1. <l-button> / <r-button>
  class LButton extends HTMLElement {
    static get observedAttributes() { return ['label', 'icon', 'severity', 'variant', 'disabled', 'size', 'type', 'loading', 'rounded', 'icon-pos']; }
    connectedCallback() { onReady(this, this.render); }
    attributeChangedCallback() { if (this._rendered) this.render(); }
    render() {
      this._rendered = true;
      const label = this.getAttribute('label') || (this._origText = (this._origText !== undefined ? this._origText : this.textContent.trim())) || '';
      const icon = this.getAttribute('icon');
      const iconPos = this.getAttribute('icon-pos') || 'left';
      const severity = this.getAttribute('severity') || this.getAttribute('variant') || 'primary';
      const size = this.getAttribute('size') || 'normal';
      const disabled = this.hasAttribute('disabled') || this.hasAttribute('loading');
      const loading = this.hasAttribute('loading');
      const rounded = this.hasAttribute('rounded');
      const iconOnly = this.hasAttribute('icon-only') || (!label && icon);
      const type = this.getAttribute('type') || 'button';

      let iconHtml = '';
      if (loading) {
        iconHtml = `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" style="animation:rikatSpin 0.7s linear infinite;"><circle cx="12" cy="12" r="10" stroke-opacity="0.25"></circle><path d="M12 2a10 10 0 0 1 10 10" stroke="currentColor"></path></svg>`;
      } else if (icon) {
        iconHtml = `<l-icon name="${icon}" size="${size === 'sm' ? 13 : size === 'lg' ? 17 : 15}"></l-icon>`;
      }

      const classes = [
        'c-btn',
        `c-btn-${severity}`,
        size === 'sm' ? 'c-btn-sm' : size === 'lg' ? 'c-btn-lg' : '',
        rounded ? 'c-btn-rounded' : '',
        iconOnly ? 'c-btn-icon-only' : ''
      ].filter(Boolean).join(' ');

      this.innerHTML = `
        <button type="${type}" class="${classes}" ${disabled ? 'disabled' : ''}>
          ${iconPos === 'left' ? iconHtml : ''}
          ${label ? `<span>${label}</span>` : ''}
          ${iconPos === 'right' ? iconHtml : ''}
        </button>
      `;
    }
  }
  defElement('l-button', LButton);
  defElement('r-button', LButton);

  // 2. <l-card> / <r-card>
  class LCard extends HTMLElement {
    static get observedAttributes() { return ['header', 'subheader']; }
    connectedCallback() { onReady(this, this.render); }
    render() {
      if (this._rendered) return;
      this._rendered = true;
      const header = this.getAttribute('header');
      const subheader = this.getAttribute('subheader');
      
      const frag = document.createDocumentFragment();
      while (this.firstChild) {
        frag.appendChild(this.firstChild);
      }

      this.innerHTML = `
        <div class="c-card">
          ${header ? `
            <div class="c-card-header" style="margin-bottom:0.75rem;">
              <div style="font-size:1.1rem;font-weight:700;color:#ffffff;">${header}</div>
              ${subheader ? `<div style="font-size:0.8rem;color:var(--text-muted);margin-top:2px;">${subheader}</div>` : ''}
            </div>
          ` : ''}
          <div class="c-card-body"></div>
        </div>
      `;
      this.querySelector('.c-card-body').appendChild(frag);
    }
  }
  defElement('l-card', LCard);
  defElement('r-card', LCard);

  // 3. <l-badge> / <l-tag>
  class LBadge extends HTMLElement {
    static get observedAttributes() { return ['value', 'severity', 'dot', 'pill', 'size']; }
    connectedCallback() { onReady(this, this.render); }
    attributeChangedCallback() { if (this._rendered) this.render(); }
    render() {
      this._rendered = true;
      const val = this.getAttribute('value') || this.textContent.trim();
      const severity = this.getAttribute('severity') || 'primary';
      const isDot = this.hasAttribute('dot');
      const isPill = this.hasAttribute('pill');
      const size = this.getAttribute('size') || 'normal';

      if (isDot) {
        this.innerHTML = `
          <span class="c-badge-dot-wrapper" style="display:inline-flex;align-items:center;gap:0.4rem;">
            <span class="c-badge-dot c-badge-dot-${severity}" style="width:8px;height:8px;border-radius:50%;background:${severity === 'success' ? '#10b981' : severity === 'danger' ? '#ef4444' : severity === 'warning' ? '#f59e0b' : '#ff2d20'};display:inline-block;box-shadow:0 0 0 2px rgba(0,0,0,0.5);"></span>
            ${val ? `<span style="font-size:0.775rem;font-weight:500;color:#f4f4f5;">${val}</span>` : ''}
          </span>
        `;
      } else {
        this.innerHTML = `<span class="c-badge c-badge-${severity} ${isPill ? 'c-badge-pill' : ''} ${size === 'lg' ? 'c-badge-lg' : ''}">${val}</span>`;
      }
    }
  }
  defElement('l-badge', LBadge);
  defElement('r-badge', LBadge);
  defElement('l-tag', LBadge);

  // 4. <l-alert> / <r-alert>
  class LAlert extends HTMLElement {
    static get observedAttributes() { return ['title', 'severity', 'icon', 'closable', 'banner']; }
    connectedCallback() { onReady(this, this.render); }
    render() {
      if (this._rendered) return;
      this._rendered = true;
      const title = this.getAttribute('title') || '';
      const severity = this.getAttribute('severity') || 'info';
      const icon = this.getAttribute('icon') || (severity === 'success' ? 'check-circle' : severity === 'danger' ? 'x-circle' : severity === 'warning' ? 'alert-circle' : 'info');
      const closable = this.hasAttribute('closable');
      const banner = this.hasAttribute('banner');
      const body = this.innerHTML;

      this.innerHTML = `
        <div class="c-alert c-alert-${severity} ${banner ? 'c-alert-banner' : ''}">
          <l-icon name="${icon}" size="18" class="c-alert-icon"></l-icon>
          <div class="c-alert-content" style="flex:1;">
            ${title ? `<div class="c-alert-title">${title}</div>` : ''}
            <div class="c-alert-desc">${body}</div>
          </div>
          ${closable ? `<button type="button" class="c-alert-close" style="all:unset;cursor:pointer;color:currentColor;opacity:0.6;padding:2px;display:flex;align-items:center;justify-content:center;" onmouseover="this.style.opacity='1'" onmouseout="this.style.opacity='0.6'"><l-icon name="x" size="14"></l-icon></button>` : ''}
        </div>
      `;

      if (closable) {
        const btn = this.querySelector('.c-alert-close');
        if (btn) {
          btn.addEventListener('click', () => {
            this.style.transition = 'opacity 0.2s ease, transform 0.2s ease';
            this.style.opacity = '0';
            this.style.transform = 'translateY(-4px)';
            setTimeout(() => this.remove(), 200);
          });
        }
      }
    }
  }
  defElement('l-alert', LAlert);
  defElement('r-alert', LAlert);

  // 5. <l-input> / <r-input>
  class LInput extends HTMLElement {
    static get observedAttributes() { return ['placeholder', 'type', 'icon', 'label', 'disabled', 'prefix', 'suffix', 'clearable', 'toggle-mask', 'invalid', 'error-message', 'value']; }
    connectedCallback() { onReady(this, this.render); }
    attributeChangedCallback() { if (this._rendered) this.render(); }
    render() {
      this._rendered = true;
      const ph = this.getAttribute('placeholder') || '';
      let type = this.getAttribute('type') || 'text';
      const icon = this.getAttribute('icon');
      const label = this.getAttribute('label');
      const prefix = this.getAttribute('prefix');
      const suffix = this.getAttribute('suffix');
      const clearable = this.hasAttribute('clearable');
      const toggleMask = this.hasAttribute('toggle-mask');
      const invalid = this.hasAttribute('invalid');
      const errorMsg = this.getAttribute('error-message') || '';
      const disabled = this.hasAttribute('disabled');
      const valModel = this.getAttribute('[(value)]') || this.getAttribute('[(model)]');
      let val = this.getAttribute('value') || '';
      if (valModel && Rikat.get(valModel) !== undefined) val = String(Rikat.get(valModel));

      const inputId = 'in_' + Math.random().toString(36).substr(2, 6);

      let leadAddon = '';
      if (prefix) leadAddon = `<span style="display:flex;align-items:center;padding:0 0.75rem;background:#202025;border:1px solid ${invalid ? '#ef4444' : '#27272a'};border-right:none;border-radius:8px 0 0 8px;font-size:0.8rem;color:#a1a1aa;font-weight:600;">${prefix}</span>`;
      else if (icon) leadAddon = `<l-icon name="${icon}" size="15" class="c-input-icon" style="position:absolute;left:0.85rem;top:50%;transform:translateY(-50%);color:${invalid ? '#ef4444' : '#71717a'};pointer-events:none;"></l-icon>`;

      let trailAddon = '';
      if (suffix) {
        trailAddon = `<span style="display:flex;align-items:center;padding:0 0.75rem;background:#202025;border:1px solid ${invalid ? '#ef4444' : '#27272a'};border-left:none;border-radius:0 8px 8px 0;font-size:0.8rem;color:#a1a1aa;font-weight:600;">${suffix}</span>`;
      } else if (toggleMask && (type === 'password' || this._isRevealed)) {
        trailAddon = `<button type="button" class="c-input-mask-toggle" style="all:unset;cursor:pointer;position:absolute;right:0.75rem;top:50%;transform:translateY(-50%);color:#71717a;display:flex;align-items:center;justify-content:center;transition:color 0.15s ease;" onmouseover="this.style.color='#f4f4f5'" onmouseout="this.style.color='#71717a'"><l-icon name="${this._isRevealed ? 'eye-off' : 'eye'}" size="15"></l-icon></button>`;
      } else if (clearable) {
        trailAddon = `<button type="button" class="c-input-clear" style="all:unset;cursor:pointer;position:absolute;right:0.75rem;top:50%;transform:translateY(-50%);color:#71717a;display:flex;align-items:center;justify-content:center;transition:color 0.15s ease;" onmouseover="this.style.color='#f4f4f5'" onmouseout="this.style.color='#71717a'"><l-icon name="x" size="13"></l-icon></button>`;
      }

      const inputStyle = [
        'width:100%',
        'height:38px',
        'background:#161619',
        `border:1px solid ${invalid ? '#ef4444' : '#27272a'}`,
        prefix && suffix ? 'border-radius:0' : prefix ? 'border-radius:0 8px 8px 0' : suffix ? 'border-radius:8px 0 0 8px' : 'border-radius:8px',
        icon ? 'padding:0 0.85rem 0 2.45rem' : (clearable || toggleMask) ? 'padding:0 2.25rem 0 0.85rem' : 'padding:0 0.85rem',
        'color:#f4f4f5',
        'font-size:0.875rem',
        'outline:none',
        'transition:all 0.15s ease',
        'font-family:inherit',
        invalid ? 'box-shadow:0 0 0 1px #ef4444' : ''
      ].filter(Boolean).join(';');

      const actualType = (toggleMask && this._isRevealed) ? 'text' : type;

      this.innerHTML = `
        <div class="c-input-group" style="width:100%;margin-bottom:0.5rem;">
          ${label ? `<label for="${inputId}" style="display:block;font-size:0.8rem;font-weight:600;color:${invalid ? '#ef4444' : '#f4f4f5'};margin-bottom:0.4rem;">${label}</label>` : ''}
          <div style="position:relative;display:flex;align-items:stretch;width:100%;">
            ${leadAddon}
            <input id="${inputId}" type="${actualType}" value="${val}" placeholder="${ph}" ${disabled ? 'disabled' : ''} style="${inputStyle}" onfocus="if(!${invalid}) { this.style.borderColor='#ff2d20';this.style.boxShadow='0 0 0 2px rgba(255,45,32,0.15)'; }" onblur="if(!${invalid}) { this.style.borderColor='#27272a';this.style.boxShadow='none'; }">
            ${trailAddon}
          </div>
          ${invalid && errorMsg ? `<span style="display:flex;align-items:center;gap:0.35rem;font-size:0.75rem;color:#ef4444;margin-top:0.35rem;"><l-icon name="alert-circle" size="13"></l-icon> ${errorMsg}</span>` : ''}
        </div>
      `;

      const inp = this.querySelector('input');
      inp.addEventListener('input', (e) => {
        this.setAttribute('value', e.target.value);
        if (valModel) Rikat.set(valModel, e.target.value);
        this.dispatchEvent(new CustomEvent('input', { detail: { value: e.target.value } }));
      });

      if (toggleMask) {
        const maskBtn = this.querySelector('.c-input-mask-toggle');
        if (maskBtn) {
          maskBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            this._isRevealed = !this._isRevealed;
            this.render();
          });
        }
      }

      if (clearable) {
        const clrBtn = this.querySelector('.c-input-clear');
        if (clrBtn) {
          clrBtn.addEventListener('click', () => {
            inp.value = '';
            this.setAttribute('value', '');
            if (valModel) Rikat.set(valModel, '');
            inp.focus();
            this.dispatchEvent(new CustomEvent('input', { detail: { value: '' } }));
          });
        }
      }
    }
  }
  defElement('l-input', LInput);
  defElement('r-input', LInput);

  // 6. <l-switch> / <r-switch>
  class LSwitch extends HTMLElement {
    static get observedAttributes() { return ['label', 'checked', 'disabled']; }
    connectedCallback() { onReady(this, this.render); }
    render() {
      this._rendered = true;
      const label = this.getAttribute('label') || '';
      const checked = this.hasAttribute('checked');
      const disabled = this.hasAttribute('disabled');
      const model = this.getAttribute('[(model)]') || '';
      this.innerHTML = `
        <label style="display:inline-flex;align-items:center;gap:0.75rem;cursor:pointer;">
          <span class="c-switch">
            <input type="checkbox" ${checked ? 'checked' : ''} ${disabled ? 'disabled' : ''} ${model ? `[(model)]="${model}"` : ''}>
            <span class="c-slider"></span>
          </span>
          ${label ? `<span style="font-size:0.85rem;color:#ffffff;font-weight:500;">${label}</span>` : ''}
        </label>
      `;
    }
  }
  defElement('l-switch', LSwitch);
  defElement('r-switch', LSwitch);

  // 7. <l-progress> / <r-progress>
  class LProgress extends HTMLElement {
    static get observedAttributes() { return ['value', 'show-value']; }
    connectedCallback() { onReady(this, this.render); }
    attributeChangedCallback() { if (this._rendered) this.render(); }
    render() {
      this._rendered = true;
      const val = parseInt(this.getAttribute('value') || '0', 10);
      const showVal = this.hasAttribute('show-value') || this.hasAttribute('showvalue');
      this.innerHTML = `
        <div style="width:100%;display:flex;align-items:center;gap:0.75rem;">
          <div class="c-progress-track" style="flex:1;">
            <div class="c-progress-bar" style="width:${val}%;"></div>
          </div>
          ${showVal ? `<span style="font-size:0.75rem;font-weight:600;color:var(--text-muted);min-width:32px;">${val}%</span>` : ''}
        </div>
      `;
    }
  }
  defElement('l-progress', LProgress);
  defElement('r-progress', LProgress);

  // 8. <l-avatar> & <l-avatar-group>
  class LAvatar extends HTMLElement {
    static get observedAttributes() { return ['label', 'image', 'shape', 'size']; }
    connectedCallback() { onReady(this, this.render); }
    render() {
      this._rendered = true;
      const label = this.getAttribute('label') || '';
      const img = this.getAttribute('image');
      const shape = this.getAttribute('shape') || 'circle';
      const size = this.getAttribute('size') || 'normal';
      const dim = size === 'large' ? '46px' : size === 'sm' ? '28px' : '36px';
      const radius = shape === 'circle' ? '50%' : '6px';
      
      if (img) {
        this.innerHTML = `<img src="${img}" class="c-avatar" style="width:${dim};height:${dim};border-radius:${radius};">`;
      } else {
        this.innerHTML = `<div class="c-avatar" style="width:${dim};height:${dim};border-radius:${radius};">${label}</div>`;
      }
    }
  }
  defElement('l-avatar', LAvatar);
  defElement('r-avatar', LAvatar);

  class LAvatarGroup extends HTMLElement {
    connectedCallback() { this.classList.add('c-avatar-group'); }
  }
  defElement('l-avatar-group', LAvatarGroup);
  defElement('r-avatar-group', LAvatarGroup);

  // 9. <l-skeleton>
  class LSkeleton extends HTMLElement {
    static get observedAttributes() { return ['width', 'height', 'shape', 'border-radius']; }
    connectedCallback() { onReady(this, this.render); }
    attributeChangedCallback() { if (this._rendered) this.render(); }
    render() {
      this._rendered = true;
      const w = this.getAttribute('width') || '100%';
      const h = this.getAttribute('height') || '1.25rem';
      const shape = this.getAttribute('shape') || 'rect';
      const customR = this.getAttribute('border-radius');
      const r = customR || (shape === 'circle' ? '50%' : '6px');

      this.style.display = shape === 'circle' ? 'inline-block' : 'block';
      this.style.width = w;
      if (shape === 'circle') this.style.height = h;

      this.innerHTML = `
        <div class="c-skeleton" style="width:100%;height:${h};border-radius:${r};background:linear-gradient(90deg, #27272e 25%, #383844 50%, #27272e 75%);background-size:200% 100%;animation:rikatShimmer 1.5s ease-in-out infinite;"></div>
      `;
    }
  }
  defElement('l-skeleton', LSkeleton);
  defElement('r-skeleton', LSkeleton);

  // 10. <l-stat> (Statistik KPI)
  class LStat extends HTMLElement {
    static get observedAttributes() { return ['label', 'value', 'trend', 'desc', 'severity']; }
    connectedCallback() { onReady(this, this.render); }
    render() {
      this._rendered = true;
      const label = this.getAttribute('label') || '';
      const value = this.getAttribute('value') || '';
      const trend = this.getAttribute('trend');
      const desc = this.getAttribute('desc');
      const severity = this.getAttribute('severity') || 'success';
      this.innerHTML = `
        <div class="c-card">
          <div style="display:flex;justify-content:space-between;align-items:flex-start;margin-bottom:0.5rem;">
            <span style="font-size:0.8rem;color:var(--text-dim);font-weight:500;">${label}</span>
            ${trend ? `<span class="c-badge c-badge-${severity}" style="font-size:0.7rem;">${trend}</span>` : ''}
          </div>
          <div style="font-size:1.75rem;font-weight:800;color:#ffffff;">${value}</div>
          ${desc ? `<div style="font-size:0.75rem;color:var(--text-muted);margin-top:0.25rem;">${desc}</div>` : ''}
        </div>
      `;
    }
  }
  defElement('l-stat', LStat);
  defElement('r-stat', LStat);

  // 11. <l-empty-state>
  class LEmptyState extends HTMLElement {
    static get observedAttributes() { return ['title', 'desc', 'icon']; }
    connectedCallback() { onReady(this, this.render); }
    render() {
      if (this._rendered) return;
      this._rendered = true;
      const title = this.getAttribute('title') || 'Belum Ada Data';
      const desc = this.getAttribute('desc') || '';
      const icon = this.getAttribute('icon') || 'inbox';
      const actions = this.innerHTML;
      this.innerHTML = `
        <div class="c-empty-state">
          <div class="c-empty-icon"><l-icon name="${icon}" size="28"></l-icon></div>
          <div class="c-empty-title">${title}</div>
          ${desc ? `<div class="c-empty-desc">${desc}</div>` : ''}
          <div style="display:flex;gap:0.75rem;justify-content:center;">${actions}</div>
        </div>
      `;
    }
  }
  defElement('l-empty-state', LEmptyState);
  defElement('r-empty-state', LEmptyState);

  // 12. <l-accordion> & <l-accordion-tab>
  class LAccordion extends HTMLElement {
    connectedCallback() { this.style.display = 'block'; }
  }
  defElement('l-accordion', LAccordion);
  defElement('r-accordion', LAccordion);

  class LAccordionTab extends HTMLElement {
    connectedCallback() { onReady(this, this.render); }
    render() {
      if (this._rendered) return;
      this._rendered = true;
      const header = this.getAttribute('header') || 'Panel';
      const isClosed = !this.hasAttribute('open');

      const frag = document.createDocumentFragment();
      while (this.firstChild) {
        frag.appendChild(this.firstChild);
      }

      this.innerHTML = `
        <div class="c-accordion-item">
          <button type="button" class="c-accordion-header">
            <span>${header}</span>
            <l-icon name="chevron-down" size="14" style="transition:transform 0.2s;transform:${isClosed ? 'none' : 'rotate(180deg)'}"></l-icon>
          </button>
          <div class="c-accordion-body" style="display:${isClosed ? 'none' : 'block'};">
          </div>
        </div>
      `;
      this.querySelector('.c-accordion-body').appendChild(frag);

      const btn = this.querySelector('.c-accordion-header');
      const body = this.querySelector('.c-accordion-body');
      const chevron = this.querySelector('l-icon');
      if (btn && body) {
        btn.addEventListener('click', () => {
          const closed = body.style.display === 'none';
          body.style.display = closed ? 'block' : 'none';
          if (chevron) chevron.style.transform = closed ? 'rotate(180deg)' : 'none';
        });
      }
    }
  }
  defElement('l-accordion-tab', LAccordionTab);
  defElement('r-accordion-tab', LAccordionTab);

  // 13. <l-tabs> & <l-tab-panel>
  class LTabs extends HTMLElement {
    connectedCallback() { onReady(this, this.render); }
    render() {
      if (this._rendered) return;
      this._rendered = true;
      const panels = Array.from(this.querySelectorAll('l-tab-panel, r-tab-panel'));
      if (panels.length === 0) return;
      const headerBar = document.createElement('div');
      headerBar.className = 'c-tabs';
      headerBar.style.marginBottom = '1.25rem';
      
      panels.forEach((p, idx) => {
        const title = p.getAttribute('header') || `Tab ${idx + 1}`;
        const btn = document.createElement('button');
        btn.type = 'button';
        btn.className = `c-tab ${idx === 0 ? 'active' : ''}`;
        btn.textContent = title;
        btn.addEventListener('click', () => {
          headerBar.querySelectorAll('.c-tab').forEach(b => b.classList.remove('active'));
          btn.classList.add('active');
          panels.forEach((panel, pIdx) => {
            panel.style.display = pIdx === idx ? 'block' : 'none';
          });
        });
        headerBar.appendChild(btn);
        p.style.display = idx === 0 ? 'block' : 'none';
      });
      
      this.insertBefore(headerBar, this.firstChild);
    }
  }
  defElement('l-tabs', LTabs);
  defElement('r-tabs', LTabs);

  class LTabPanel extends HTMLElement {}
  defElement('l-tab-panel', LTabPanel);
  defElement('r-tab-panel', LTabPanel);

  // 14. <l-breadcrumb>
  class LBreadcrumb extends HTMLElement {
    static get observedAttributes() { return ['items']; }
    connectedCallback() { onReady(this, this.render); }
    attributeChangedCallback() { if (this._rendered) this.render(); }
    render() {
      this._rendered = true;
      const raw = this.getAttribute('items');
      let items = [];
      if (raw) {
        try { items = JSON.parse(raw); } catch(e) {}
      }
      
      const nav = document.createElement('nav');
      nav.className = 'c-breadcrumb';
      nav.style.cssText = 'display:flex;align-items:center;gap:0.4rem;font-size:0.875rem;flex-wrap:wrap;';

      if (items.length > 0) {
        items.forEach((item, idx) => {
          const isLast = (idx === items.length - 1);
          if (item.url && !isLast) {
            const a = document.createElement('a');
            a.href = item.url;
            a.style.cssText = 'color:#a1a1aa;text-decoration:none;display:inline-flex;align-items:center;gap:0.35rem;transition:color 0.15s ease;';
            a.onmouseover = () => a.style.color = '#f4f4f5';
            a.onmouseout = () => a.style.color = '#a1a1aa';
            if (item.icon) a.innerHTML = `<l-icon name="${item.icon}" size="13"></l-icon> <span>${item.label}</span>`;
            else a.textContent = item.label;
            nav.appendChild(a);
          } else {
            const s = document.createElement('span');
            s.style.cssText = 'color:#f4f4f5;font-weight:600;display:inline-flex;align-items:center;gap:0.35rem;';
            if (item.icon) s.innerHTML = `<l-icon name="${item.icon}" size="13"></l-icon> <span>${item.label}</span>`;
            else s.textContent = item.label;
            nav.appendChild(s);
          }
          if (!isLast) {
            const sep = document.createElement('span');
            sep.style.cssText = 'color:#52525b;display:inline-flex;align-items:center;';
            sep.innerHTML = '<l-icon name="chevron-right" size="13"></l-icon>';
            nav.appendChild(sep);
          }
        });
      } else {
        const children = Array.from(this.querySelectorAll('a, span'));
        children.forEach((child, idx) => {
          nav.appendChild(child);
          if (idx < children.length - 1) {
            const sep = document.createElement('span');
            sep.style.cssText = 'color:#52525b;display:inline-flex;align-items:center;';
            sep.innerHTML = '<l-icon name="chevron-right" size="13"></l-icon>';
            nav.appendChild(sep);
          }
        });
      }

      this.innerHTML = '';
      this.appendChild(nav);
    }
  }
  defElement('l-breadcrumb', LBreadcrumb);
  defElement('r-breadcrumb', LBreadcrumb);

  // 15. <l-stepper> & <l-step>
  class LStepper extends HTMLElement {
    connectedCallback() { onReady(this, this.render); }
    render() {
      if (this._rendered) return;
      this._rendered = true;
      const steps = Array.from(this.querySelectorAll('l-step, r-step'));
      const container = document.createElement('div');
      container.className = 'c-stepper';
      steps.forEach((s, idx) => {
        const label = s.getAttribute('label') || `Langkah ${idx + 1}`;
        const status = s.getAttribute('status') || (idx === 0 ? 'active' : 'pending');
        const item = document.createElement('div');
        item.className = 'c-step-item';
        
        let iconOrNum = idx + 1;
        if (status === 'completed') iconOrNum = '<l-icon name="check" size="14"></l-icon>';
        
        item.innerHTML = `
          <div class="c-step-bubble ${status}">${iconOrNum}</div>
          <span class="c-step-label ${status === 'pending' ? 'text-dim' : ''}">${label}</span>
        `;
        container.appendChild(item);
        if (idx < steps.length - 1) {
          const line = document.createElement('div');
          line.className = `c-step-line ${status === 'completed' ? 'completed' : ''}`;
          container.appendChild(line);
        }
      });
      this.innerHTML = '';
      this.appendChild(container);
      this.dispatchEvent(new CustomEvent('rikat-tree-ready', { bubbles: true }));
    }
  }
  defElement('l-stepper', LStepper);
  defElement('r-stepper', LStepper);

  class LStep extends HTMLElement {}
  defElement('l-step', LStep);
  defElement('r-step', LStep);

  // 16. <l-timeline> & <l-timeline-event>
  class LTimeline extends HTMLElement {
    connectedCallback() { onReady(this, this.render); }
    render() {
      if (this._rendered) return;
      this._rendered = true;
      const events = Array.from(this.querySelectorAll('l-timeline-event, r-timeline-event'));
      const container = document.createElement('div');
      container.className = 'c-timeline';
      events.forEach(e => {
        const title = e.getAttribute('title') || '';
        const time = e.getAttribute('time') || '';
        const dotColor = e.getAttribute('color') || 'var(--primary)';
        const desc = e.innerHTML;
        const item = document.createElement('div');
        item.className = 'c-timeline-item';
        item.innerHTML = `
          <div class="c-timeline-dot" style="background-color:${dotColor};"></div>
          ${time ? `<div class="c-timeline-time">${time}</div>` : ''}
          <div class="c-timeline-title">${title}</div>
          <div class="c-timeline-desc">${desc}</div>
        `;
        container.appendChild(item);
      });
      this.innerHTML = '';
      this.appendChild(container);
    }
  }
  defElement('l-timeline', LTimeline);
  defElement('r-timeline', LTimeline);

  class LTimelineEvent extends HTMLElement {}
  defElement('l-timeline-event', LTimelineEvent);
  defElement('r-timeline-event', LTimelineEvent);


  // ==========================================================================
  // New Enterprise Components (Slider, Rating, Segmented, Divider, Chip, Kbd,
  // Otp, ColorPicker, Tree, Drawer)
  // ==========================================================================

  // 17. <l-slider> & <r-slider>
  class LSlider extends HTMLElement {
    static get observedAttributes() { return ['min', 'max', 'step', 'value', 'disabled']; }
    connectedCallback() { onReady(this, this.render); }
    attributeChangedCallback() { if (this._rendered) this.render(); }
    render() {
      this._rendered = true;
      const min = parseFloat(this.getAttribute('min') || '0');
      const max = parseFloat(this.getAttribute('max') || '100');
      const step = parseFloat(this.getAttribute('step') || '1');
      let val = parseFloat(this.getAttribute('value') || String(min));
      const disabled = this.hasAttribute('disabled');
      const valModel = this.getAttribute('[(value)]') || this.getAttribute('[(model)]');
      if (valModel && Rikat.get(valModel) !== undefined) {
        val = parseFloat(Rikat.get(valModel));
      }
      const pct = Math.max(0, Math.min(100, ((val - min) / (max - min)) * 100));

      this.innerHTML = `
        <div class="c-slider-wrapper" style="display:flex;align-items:center;gap:1rem;width:100%;">
          <div class="c-slider-track" style="position:relative;flex:1;height:6px;background:#27272a;border-radius:3px;cursor:${disabled ? 'not-allowed' : 'pointer'};">
            <div class="c-slider-fill" style="position:absolute;left:0;top:0;height:100%;width:${pct}%;background:#ff2d20;border-radius:3px;"></div>
            <div class="c-slider-thumb" style="position:absolute;left:${pct}%;top:50%;transform:translate(-50%, -50%);width:16px;height:16px;background:#ffffff;border:2px solid #ff2d20;border-radius:50%;box-shadow:0 2px 6px rgba(0,0,0,0.4);transition:transform 0.1s ease;"></div>
          </div>
          <span class="c-slider-val" style="font-size:0.85rem;font-weight:600;color:#f4f4f5;min-width:32px;text-align:right;">${val}</span>
        </div>
      `;

      if (!disabled) {
        const track = this.querySelector('.c-slider-track');
        const updateFromPos = (clientX) => {
          const rect = track.getBoundingClientRect();
          const pos = Math.max(0, Math.min(1, (clientX - rect.left) / rect.width));
          let newVal = min + Math.round((pos * (max - min)) / step) * step;
          newVal = Math.max(min, Math.min(max, newVal));
          this.setAttribute('value', String(newVal));
          if (valModel) Rikat.set(valModel, newVal);
          this.dispatchEvent(new CustomEvent('change', { detail: { value: newVal } }));
          this.render();
        };

        let dragging = false;
        track.addEventListener('mousedown', (e) => {
          dragging = true;
          updateFromPos(e.clientX);
          const onMove = (ev) => { if (dragging) updateFromPos(ev.clientX); };
          const onUp = () => { dragging = false; window.removeEventListener('mousemove', onMove); window.removeEventListener('mouseup', onUp); };
          window.addEventListener('mousemove', onMove);
          window.addEventListener('mouseup', onUp);
        });
      }
    }
  }
  defElement('l-slider', LSlider);
  defElement('r-slider', LSlider);

  // 18. <l-rating> & <r-rating>
  class LRating extends HTMLElement {
    static get observedAttributes() { return ['value', 'stars', 'readonly']; }
    connectedCallback() { onReady(this, this.render); }
    attributeChangedCallback() { if (this._rendered) this.render(); }
    render() {
      this._rendered = true;
      const stars = parseInt(this.getAttribute('stars') || '5', 10);
      let val = parseInt(this.getAttribute('value') || '0', 10);
      const readonly = this.hasAttribute('readonly');
      const valModel = this.getAttribute('[(value)]') || this.getAttribute('[(model)]');
      if (valModel && Rikat.get(valModel) !== undefined) {
        val = parseInt(Rikat.get(valModel), 10);
      }

      let starsHtml = '';
      for (let i = 1; i <= stars; i++) {
        const filled = i <= val;
        starsHtml += `
          <span class="c-rating-star" data-star="${i}" style="cursor:${readonly ? 'default' : 'pointer'};display:inline-flex;padding:2px;color:${filled ? '#fbbf24' : '#3f3f46'};transition:all 0.15s ease;">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="${filled ? 'currentColor' : 'none'}" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"></polygon>
            </svg>
          </span>
        `;
      }

      this.innerHTML = `<div class="c-rating" style="display:inline-flex;align-items:center;gap:0.25rem;">${starsHtml}</div>`;

      if (!readonly) {
        this.querySelectorAll('.c-rating-star').forEach(starEl => {
          starEl.addEventListener('click', () => {
            const star = parseInt(starEl.getAttribute('data-star'), 10);
            const newVal = star === val ? 0 : star;
            this.setAttribute('value', String(newVal));
            if (valModel) Rikat.set(valModel, newVal);
            this.dispatchEvent(new CustomEvent('change', { detail: { value: newVal } }));
            this.render();
          });
          starEl.addEventListener('mouseenter', () => {
            const star = parseInt(starEl.getAttribute('data-star'), 10);
            this.querySelectorAll('.c-rating-star').forEach((s, idx) => {
              s.style.color = (idx + 1 <= star) ? '#fbbf24' : '#3f3f46';
            });
          });
        });
        this.querySelector('.c-rating').addEventListener('mouseleave', () => {
          this.querySelectorAll('.c-rating-star').forEach((s, idx) => {
            s.style.color = (idx + 1 <= val) ? '#fbbf24' : '#3f3f46';
          });
        });
      }
    }
  }
  defElement('l-rating', LRating);
  defElement('r-rating', LRating);

  // 19. <l-segmented> & <r-segmented>
  class LSegmented extends HTMLElement {
    static get observedAttributes() { return ['options', 'value']; }
    connectedCallback() { onReady(this, this.render); }
    attributeChangedCallback() { if (this._rendered) this.render(); }
    render() {
      this._rendered = true;
      const optsStr = this.getAttribute('options') || '';
      const opts = optsStr.split(',').map(s => s.trim()).filter(Boolean);
      const valModel = this.getAttribute('[(value)]') || this.getAttribute('[(model)]');
      let curVal = this.getAttribute('value') || (opts[0] || '');
      if (valModel && Rikat.get(valModel) !== undefined) {
        curVal = String(Rikat.get(valModel));
      }

      this.innerHTML = `
        <div class="c-segmented" style="display:inline-flex;background:#18181b;border:1px solid #27272a;border-radius:8px;padding:3px;gap:2px;">
          ${opts.map(opt => {
            const active = opt === curVal;
            return `
              <button type="button" class="c-segmented-item ${active ? 'active' : ''}" data-opt="${opt}" style="all:unset;cursor:pointer;padding:0.4rem 0.9rem;font-size:0.8rem;font-weight:${active ? '600' : '500'};color:${active ? '#ffffff' : '#a1a1aa'};background:${active ? '#27272a' : 'transparent'};border-radius:6px;transition:all 0.15s ease;text-align:center;">
                ${opt}
              </button>
            `;
          }).join('')}
        </div>
      `;

      this.querySelectorAll('.c-segmented-item').forEach(btn => {
        btn.addEventListener('click', () => {
          const opt = btn.getAttribute('data-opt');
          this.setAttribute('value', opt);
          if (valModel) Rikat.set(valModel, opt);
          this.dispatchEvent(new CustomEvent('change', { detail: { value: opt } }));
          this.render();
        });
      });
    }
  }
  defElement('l-segmented', LSegmented);
  defElement('r-segmented', LSegmented);

  // 20. <l-divider> & <r-divider>
  class LDivider extends HTMLElement {
    connectedCallback() { onReady(this, this.render); }
    render() {
      if (this._rendered) return;
      this._rendered = true;
      const layout = this.getAttribute('layout') || 'horizontal';
      const align = this.getAttribute('align') || 'center';
      const label = this.getAttribute('label') || this.textContent.trim() || '';

      if (layout === 'vertical') {
        this.innerHTML = `<div class="c-divider c-divider-vertical" style="display:inline-block;width:1px;height:100%;min-height:24px;background:#27272a;margin:0 1rem;vertical-align:middle;"></div>`;
      } else {
        if (!label) {
          this.innerHTML = `<div class="c-divider" style="width:100%;height:1px;background:#27272a;margin:1.25rem 0;"></div>`;
        } else {
          this.innerHTML = `
            <div class="c-divider-with-text" style="display:flex;align-items:center;width:100%;margin:1.25rem 0;color:#71717a;font-size:0.75rem;font-weight:600;text-transform:uppercase;letter-spacing:0.05em;">
              <span style="flex:${align === 'right' ? '1' : align === 'center' ? '1' : '0 0 1rem'};height:1px;background:#27272a;"></span>
              <span style="padding:0 0.85rem;white-space:nowrap;">${label}</span>
              <span style="flex:${align === 'left' ? '1' : align === 'center' ? '1' : '0 0 1rem'};height:1px;background:#27272a;"></span>
            </div>
          `;
        }
      }
    }
  }
  defElement('l-divider', LDivider);
  defElement('r-divider', LDivider);

  // 21. <l-chip> & <r-chip>
  class LChip extends HTMLElement {
    static get observedAttributes() { return ['label', 'icon', 'removable', 'image']; }
    connectedCallback() { onReady(this, this.render); }
    attributeChangedCallback() { if (this._rendered) this.render(); }
    render() {
      this._rendered = true;
      const label = this.getAttribute('label') || this.textContent.trim() || '';
      const icon = this.getAttribute('icon');
      const image = this.getAttribute('image');
      const removable = this.hasAttribute('removable');

      let leadHtml = '';
      if (image) leadHtml = `<img src="${image}" style="width:18px;height:18px;border-radius:50%;object-fit:cover;margin-left:-2px;" />`;
      else if (icon) leadHtml = `<l-icon name="${icon}" size="14"></l-icon>`;

      const removeBtn = removable ? `
        <button type="button" class="c-chip-remove" style="all:unset;cursor:pointer;color:#71717a;display:inline-flex;align-items:center;justify-content:center;margin-left:0.25rem;transition:color 0.15s ease;" onmouseover="this.style.color='#f4f4f5'" onmouseout="this.style.color='#71717a'">
          <l-icon name="x" size="12"></l-icon>
        </button>
      ` : '';

      this.innerHTML = `
        <span class="c-chip" style="display:inline-flex;align-items:center;gap:0.4rem;background:#212126;border:1px solid #2d2d34;color:#f4f4f5;font-size:0.775rem;font-weight:500;padding:0.25rem 0.65rem;border-radius:9999px;transition:all 0.15s ease;">
          ${leadHtml}
          <span>${label}</span>
          ${removeBtn}
        </span>
      `;

      if (removable) {
        const btn = this.querySelector('.c-chip-remove');
        if (btn) {
          btn.addEventListener('click', (e) => {
            e.stopPropagation();
            this.dispatchEvent(new CustomEvent('remove'));
            this.style.transition = 'opacity 0.2s, transform 0.2s';
            this.style.opacity = '0';
            this.style.transform = 'scale(0.8)';
            setTimeout(() => this.remove(), 200);
          });
        }
      }
    }
  }
  defElement('l-chip', LChip);
  defElement('r-chip', LChip);

  // 22. <l-kbd> & <r-kbd>
  class LKbd extends HTMLElement {
    connectedCallback() {
      this.style.display = 'inline-flex';
      this.style.alignItems = 'center';
      this.style.justifyContent = 'center';
      this.classList.add('c-kbd');
    }
  }
  defElement('l-kbd', LKbd);
  defElement('r-kbd', LKbd);

  // 23. <l-otp> & <r-otp>
  class LOtp extends HTMLElement {
    static get observedAttributes() { return ['length', 'disabled', 'mask']; }
    connectedCallback() { onReady(this, this.render); }
    render() {
      if (this._rendered) return;
      this._rendered = true;
      const len = parseInt(this.getAttribute('length') || '4', 10);
      const mask = this.hasAttribute('mask');
      const disabled = this.hasAttribute('disabled');
      const valModel = this.getAttribute('[(value)]') || this.getAttribute('[(model)]');

      let inputsHtml = '';
      for (let i = 0; i < len; i++) {
        inputsHtml += `
          <input type="${mask ? 'password' : 'text'}" maxlength="1" inputmode="numeric" pattern="[0-9]*" class="c-otp-digit" data-idx="${i}" ${disabled ? 'disabled' : ''} style="width:42px;height:46px;text-align:center;font-size:1.15rem;font-weight:700;color:#ffffff;background:#161619;border:1px solid #27272a;border-radius:8px;outline:none;transition:border-color 0.15s ease, box-shadow 0.15s ease;" onfocus="this.style.borderColor='#ff2d20';this.style.boxShadow='0 0 0 2px rgba(255,45,32,0.2)'" onblur="this.style.borderColor='#27272a';this.style.boxShadow='none'">
        `;
      }

      this.innerHTML = `<div class="c-otp-wrapper" style="display:inline-flex;gap:0.5rem;align-items:center;">${inputsHtml}</div>`;

      const inputs = Array.from(this.querySelectorAll('.c-otp-digit'));
      const updateVal = () => {
        const full = inputs.map(i => i.value).join('');
        this.setAttribute('value', full);
        if (valModel) Rikat.set(valModel, full);
        this.dispatchEvent(new CustomEvent('change', { detail: { value: full } }));
        if (full.length === len) {
          this.dispatchEvent(new CustomEvent('complete', { detail: { value: full } }));
        }
      };

      inputs.forEach((input, idx) => {
        input.addEventListener('input', (e) => {
          const v = e.target.value.replace(/[^0-9a-zA-Z]/g, '');
          e.target.value = v.slice(-1);
          if (e.target.value && idx < len - 1) {
            inputs[idx + 1].focus();
            inputs[idx + 1].select();
          }
          updateVal();
        });
        input.addEventListener('keydown', (e) => {
          if (e.key === 'Backspace' && !input.value && idx > 0) {
            inputs[idx - 1].focus();
          }
        });
      });
    }
  }
  defElement('l-otp', LOtp);
  defElement('r-otp', LOtp);

  // 24. <l-color-picker> & <r-color-picker>
  class LColorPicker extends HTMLElement {
    static get observedAttributes() { return ['value']; }
    connectedCallback() { onReady(this, this.render); }
    attributeChangedCallback() { if (this._rendered) this.render(); }
    render() {
      this._rendered = true;
      const valModel = this.getAttribute('[(value)]') || this.getAttribute('[(model)]');
      let curColor = this.getAttribute('value') || '#ff2d20';
      if (valModel && Rikat.get(valModel) !== undefined) {
        curColor = String(Rikat.get(valModel));
      }
      const presets = ['#ff2d20', '#f97316', '#fbbf24', '#10b981', '#06b6d4', '#3b82f6', '#8b5cf6', '#ec4899', '#ffffff'];

      this.innerHTML = `
        <div class="c-colorpicker-wrapper" style="display:inline-flex;align-items:center;gap:0.75rem;background:#18181b;border:1px solid #27272a;border-radius:8px;padding:0.5rem 0.75rem;">
          <label style="cursor:pointer;position:relative;display:inline-block;width:28px;height:28px;border-radius:6px;background:${curColor};border:2px solid #3f3f46;overflow:hidden;flex-shrink:0;">
            <input type="color" value="${curColor}" class="c-colorpicker-native" style="position:absolute;left:-20px;top:-20px;width:80px;height:80px;opacity:0;cursor:pointer;">
          </label>
          <div style="display:flex;gap:4px;flex-wrap:wrap;">
            ${presets.map(c => `
              <span class="c-color-swatch" data-color="${c}" style="cursor:pointer;width:18px;height:18px;border-radius:4px;background:${c};border:1px solid ${c === curColor ? '#ffffff' : 'rgba(255,255,255,0.15)'};transform:${c === curColor ? 'scale(1.15)' : 'scale(1)'};transition:all 0.15s ease;"></span>
            `).join('')}
          </div>
          <span class="c-color-hex" style="font-size:0.8rem;font-weight:600;font-family:monospace;color:#f4f4f5;min-width:60px;">${curColor}</span>
        </div>
      `;

      const native = this.querySelector('.c-colorpicker-native');
      native.addEventListener('input', (e) => {
        const c = e.target.value;
        this.setAttribute('value', c);
        if (valModel) Rikat.set(valModel, c);
        this.render();
      });

      this.querySelectorAll('.c-color-swatch').forEach(sw => {
        sw.addEventListener('click', () => {
          const c = sw.getAttribute('data-color');
          this.setAttribute('value', c);
          if (valModel) Rikat.set(valModel, c);
          this.render();
        });
      });
    }
  }
  defElement('l-color-picker', LColorPicker);
  defElement('r-color-picker', LColorPicker);

  // 25. <l-tree> & <r-tree>
  class LTree extends HTMLElement {
    connectedCallback() { onReady(this, this.render); }
    render() {
      if (this._rendered) return;
      this._rendered = true;
      const directNodes = Array.from(this.children).filter(c => {
        const tag = c.tagName.toLowerCase();
        return tag === 'l-tree-node' || tag === 'r-tree-node';
      });

      const container = document.createElement('div');
      container.className = 'c-tree';
      container.style.cssText = 'font-size:0.85rem;line-height:1.6;width:100%;';

      const curPath = window.location.pathname;
      const isRouteActive = (href) => {
        if (!href) return false;
        return href === curPath || (curPath === '/' && href === '/dok/changelog');
      };

      const hasActiveDescendant = (node) => {
        const href = node.getAttribute('href') || node.getAttribute('url');
        if (isRouteActive(href)) return true;
        const children = Array.from(node.children).filter(c => {
          const tag = c.tagName.toLowerCase();
          return tag === 'l-tree-node' || tag === 'r-tree-node';
        });
        return children.some(hasActiveDescendant);
      };

      const renderNode = (el) => {
        const label = el.getAttribute('label') || 'Item';
        const href = el.getAttribute('href') || el.getAttribute('url');
        const childNodes = Array.from(el.children).filter(c => {
          const tag = c.tagName.toLowerCase();
          return tag === 'l-tree-node' || tag === 'r-tree-node';
        });
        const hasChildren = childNodes.length > 0;
        const defaultIcon = hasChildren ? 'folder' : 'file';
        const icon = el.getAttribute('icon') || defaultIcon;
        const isActive = isRouteActive(href);
        const containsActive = hasActiveDescendant(el);
        const isExpanded = containsActive || (el.hasAttribute('expanded') && el.getAttribute('expanded') !== 'false');

        const item = document.createElement('div');
        item.className = 'c-tree-item';
        item.style.cssText = 'margin:1px 0;';

        const isLink = Boolean(href && !hasChildren);
        const row = document.createElement(isLink ? 'a' : 'div');
        row.className = `c-tree-row ${isActive ? 'active' : ''}`;
        if (isLink) {
          row.href = href;
        }

        const activeStyles = isActive
          ? 'background:rgba(239,68,68,0.12);color:#f87171;font-weight:600;border-left:2px solid #ef4444;'
          : 'color:#d4d4d8;';

        row.style.cssText = `display:flex;align-items:center;gap:0.45rem;padding:0.32rem 0.55rem;border-radius:6px;cursor:pointer;text-decoration:none;transition:all 0.15s ease;user-select:none;${activeStyles}`;

        row.onmouseover = () => {
          if (!isActive) {
            row.style.background = 'rgba(255,255,255,0.06)';
            row.style.color = '#ffffff';
          }
        };
        row.onmouseout = () => {
          if (!isActive) {
            row.style.background = 'transparent';
            row.style.color = '#d4d4d8';
          }
        };

        let chevronHtml = hasChildren
          ? `<l-icon name="chevron-down" size="13" style="transition:transform 0.2s;transform:${isExpanded ? 'rotate(0deg)' : 'rotate(-90deg)'};color:#a1a1aa;flex-shrink:0;"></l-icon>`
          : '<span style="width:13px;display:inline-block;flex-shrink:0;"></span>';

        const iconColor = isActive ? '#ef4444' : (hasChildren ? '#fbbf24' : '#60a5fa');

        row.innerHTML = `
          ${chevronHtml}
          <l-icon name="${icon}" size="15" color="${iconColor}"></l-icon>
          <span style="font-weight:${isActive ? '600' : '500'};line-height:1.35;">${label}</span>
        `;
        item.appendChild(row);

        if (hasChildren) {
          const childContainer = document.createElement('div');
          childContainer.className = 'c-tree-branch';
          childContainer.style.cssText = `padding-left:0.9rem;border-left:1px dashed #27272a;margin-left:0.65rem;display:${isExpanded ? 'block' : 'none'};`;
          childNodes.forEach(c => childContainer.appendChild(renderNode(c)));
          item.appendChild(childContainer);

          row.addEventListener('click', (e) => {
            if (e.target.closest('a') && e.target.closest('a') !== row) return;
            e.stopPropagation();
            const isHidden = childContainer.style.display === 'none';
            childContainer.style.display = isHidden ? 'block' : 'none';
            const chev = row.querySelector('l-icon[name="chevron-down"]');
            if (chev) {
              chev.style.transform = isHidden ? 'rotate(0deg)' : 'rotate(-90deg)';
            }
          });
        }
        return item;
      };

      directNodes.forEach(n => container.appendChild(renderNode(n)));
      this.innerHTML = '';
      this.appendChild(container);
    }
  }
  defElement('l-tree', LTree);
  defElement('r-tree', LTree);

  class LTreeNode extends HTMLElement {}
  defElement('l-tree-node', LTreeNode);
  defElement('r-tree-node', LTreeNode);

  // 26. <l-drawer> & <r-drawer>
  class LDrawer extends HTMLElement {
    connectedCallback() {
      this.visibleKey = this.getAttribute('[(visible)]') || this.getAttribute('[visible]') || this.getAttribute('visible');
      this.headerText = this.getAttribute('header') || 'Panel Samping';
      this.position = this.getAttribute('position') || 'right';
      this.style.display = 'contents';
      
      if (document.readyState === 'loading') {
        const handler = () => {
          document.removeEventListener('DOMContentLoaded', handler);
          this.initDrawer();
        };
        document.addEventListener('DOMContentLoaded', handler);
      } else {
        setTimeout(() => this.initDrawer(), 0);
      }
    }

    initDrawer() {
      if (this._initialized) return;
      this._initialized = true;

      const frag = document.createDocumentFragment();
      while (this.firstChild) {
        frag.appendChild(this.firstChild);
      }

      const isRight = this.position === 'right';
      const isLeft = this.position === 'left';
      const isBottom = this.position === 'bottom';

      this.innerHTML = `
        <div class="rikat-drawer-backdrop" style="display:none;position:fixed;inset:0;background:rgba(0,0,0,0.7);backdrop-filter:blur(4px);z-index:9999;align-items:${isBottom ? 'flex-end' : 'stretch'};justify-content:${isRight ? 'flex-end' : isLeft ? 'flex-start' : 'center'};">
          <div class="rikat-drawer-panel" style="background:#161619;border-${isRight ? 'left' : isLeft ? 'right' : 'top'}:1px solid #27272a;width:${isBottom ? '100%' : '380px'};max-width:100%;height:${isBottom ? '340px' : '100%'};box-shadow:0 0 40px rgba(0,0,0,0.8);display:flex;flex-direction:column;animation:rikatSlideIn 0.25s ease;">
            <div style="display:flex;align-items:center;justify-content:space-between;padding:1.15rem 1.35rem;border-bottom:1px solid #27272a;background:#141417;">
              <span style="font-weight:700;font-size:1rem;color:#ffffff;">${this.headerText}</span>
              <button type="button" class="rikat-drawer-close" style="all:unset;cursor:pointer;color:#a1a1aa;display:flex;align-items:center;justify-content:center;width:28px;height:28px;border-radius:6px;font-size:1.25rem;" onmouseover="this.style.color='#fff'" onmouseout="this.style.color='#a1a1aa'">&times;</button>
            </div>
            <div class="rikat-drawer-body" style="padding:1.35rem;flex:1;overflow-y:auto;color:#d4d4d8;">
            </div>
          </div>
        </div>
      `;

      this.querySelector('.rikat-drawer-body').appendChild(frag);

      this.querySelector('.rikat-drawer-close').addEventListener('click', () => this.close());
      const backdrop = this.querySelector('.rikat-drawer-backdrop');
      backdrop.addEventListener('click', (e) => {
        if (e.target === backdrop) this.close();
      });

      this.onStateUpdate();
    }

    close() {
      if (this.visibleKey) Rikat.set(this.visibleKey, false);
      else {
        const bd = this.querySelector('.rikat-drawer-backdrop');
        if (bd) bd.style.display = 'none';
        document.body.style.overflow = '';
      }
    }

    onStateUpdate() {
      if (!this.visibleKey) return;
      const isVisible = Boolean(Rikat.get(this.visibleKey));
      const bd = this.querySelector('.rikat-drawer-backdrop');
      if (bd) {
        bd.style.display = isVisible ? 'flex' : 'none';
        document.body.style.overflow = isVisible ? 'hidden' : '';
      }
    }
  }
  defElement('l-drawer', LDrawer);
  defElement('r-drawer', LDrawer);



  // ==========================================================================
  // Additional Consistent L-* Web Components (Checkbox, Radio, Textarea,
  // Dropdown, Pagination, Toast, Upload, ButtonGroup)
  // ==========================================================================

  // LCheckbox
  class LCheckbox extends HTMLElement {
    static get observedAttributes() { return ['label', 'checked', 'disabled', 'severity', 'value']; }
    connectedCallback() { onReady(this, this.render); }
    attributeChangedCallback() { if (this._rendered) this.render(); }
    render() {
      this._rendered = true;
      const label = this.getAttribute('label') || this.textContent.trim();
      const isChecked = this.hasAttribute('checked');
      const disabled = this.hasAttribute('disabled');
      const severity = this.getAttribute('severity') || 'primary';
      const valModel = this.getAttribute('[(value)]') || this.getAttribute('[(checked)]');
      let checked = isChecked;
      if (valModel && Rikat.get(valModel) !== undefined) {
        checked = Boolean(Rikat.get(valModel));
      }
      const color = severity === 'danger' ? '#ef4444' : severity === 'success' ? '#10b981' : severity === 'warning' ? '#f59e0b' : '#ff2d20';

      this.innerHTML = `
        <label class="c-checkbox-wrapper" style="display:inline-flex;align-items:center;gap:0.6rem;cursor:${disabled ? 'not-allowed' : 'pointer'};opacity:${disabled ? '0.5' : '1'};user-select:none;font-size:0.875rem;color:#f4f4f5;font-weight:500;">
          <span class="c-checkbox-box" style="width:18px;height:18px;border-radius:4px;border:1.5px solid ${checked ? color : '#3f3f46'};background:${checked ? color : '#161619'};display:inline-flex;align-items:center;justify-content:center;transition:all 0.15s ease;">
            ${checked ? '<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="#ffffff" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"></polyline></svg>' : ''}
          </span>
          ${label ? `<span>${label}</span>` : ''}
        </label>
      `;

      if (!disabled) {
        const lbl = this.querySelector('label');
        lbl.addEventListener('click', (e) => {
          e.preventDefault();
          const next = !checked;
          if (next) this.setAttribute('checked', '');
          else this.removeAttribute('checked');
          if (valModel) Rikat.set(valModel, next);
          this.dispatchEvent(new CustomEvent('change', { detail: { checked: next } }));
          this.render();
        });
      }
    }
  }
  defElement('l-checkbox', LCheckbox);
  defElement('r-checkbox', LCheckbox);

  // LRadio
  class LRadio extends HTMLElement {
    static get observedAttributes() { return ['label', 'name', 'value', 'checked', 'disabled']; }
    connectedCallback() { onReady(this, this.render); }
    attributeChangedCallback() { if (this._rendered) this.render(); }
    render() {
      this._rendered = true;
      const label = this.getAttribute('label') || this.textContent.trim();
      const name = this.getAttribute('name') || '';
      const val = this.getAttribute('value') || '';
      const isChecked = this.hasAttribute('checked');
      const disabled = this.hasAttribute('disabled');
      const valModel = this.getAttribute('[(value)]');
      let checked = isChecked;
      if (valModel && Rikat.get(valModel) !== undefined) {
        checked = (Rikat.get(valModel) === val);
      }

      this.innerHTML = `
        <label class="c-radio-wrapper" style="display:inline-flex;align-items:center;gap:0.6rem;cursor:${disabled ? 'not-allowed' : 'pointer'};opacity:${disabled ? '0.5' : '1'};user-select:none;font-size:0.875rem;color:#f4f4f5;font-weight:500;">
          <span class="c-radio-circle" style="width:18px;height:18px;border-radius:50%;border:1.5px solid ${checked ? '#ff2d20' : '#3f3f46'};background:#161619;display:inline-flex;align-items:center;justify-content:center;transition:all 0.15s ease;">
            ${checked ? '<span style="width:8px;height:8px;border-radius:50%;background:#ff2d20;"></span>' : ''}
          </span>
          ${label ? `<span>${label}</span>` : ''}
        </label>
      `;

      if (!disabled) {
        const lbl = this.querySelector('label');
        lbl.addEventListener('click', (e) => {
          e.preventDefault();
          if (name) {
            document.querySelectorAll(`l-radio[name="${name}"], r-radio[name="${name}"]`).forEach(r => {
              if (r !== this) {
                r.removeAttribute('checked');
                if (r.render) r.render();
              }
            });
          }
          this.setAttribute('checked', '');
          if (valModel) Rikat.set(valModel, val);
          this.dispatchEvent(new CustomEvent('change', { detail: { value: val } }));
          this.render();
        });
      }
    }
  }
  defElement('l-radio', LRadio);
  defElement('r-radio', LRadio);

  // LTextarea
  class LTextarea extends HTMLElement {
    static get observedAttributes() { return ['label', 'placeholder', 'rows', 'disabled', 'value', 'char-limit']; }
    connectedCallback() { onReady(this, this.render); }
    attributeChangedCallback() { if (this._rendered) this.render(); }
    render() {
      this._rendered = true;
      const label = this.getAttribute('label');
      const ph = this.getAttribute('placeholder') || '';
      const rows = this.getAttribute('rows') || '3';
      const disabled = this.hasAttribute('disabled');
      const charLimit = this.getAttribute('char-limit');
      let val = this.getAttribute('value') || this.textContent.trim();
      const valModel = this.getAttribute('[(value)]') || this.getAttribute('[(model)]');
      if (valModel && Rikat.get(valModel) !== undefined) val = String(Rikat.get(valModel));

      this.innerHTML = `
        <div class="c-textarea-group" style="width:100%;margin-bottom:0.5rem;">
          ${label ? `<label style="display:block;font-size:0.8rem;font-weight:600;color:#f4f4f5;margin-bottom:0.4rem;">${label}</label>` : ''}
          <textarea placeholder="${ph}" rows="${rows}" ${disabled ? 'disabled' : ''} style="width:100%;padding:0.75rem 0.85rem;background:#161619;border:1px solid #27272a;border-radius:8px;color:#f4f4f5;font-size:0.875rem;font-family:inherit;outline:none;resize:vertical;min-height:70px;transition:all 0.15s ease;" onfocus="this.style.borderColor='#ff2d20';this.style.boxShadow='0 0 0 2px rgba(255,45,32,0.15)'" onblur="this.style.borderColor='#27272a';this.style.boxShadow='none'">${val}</textarea>
          ${charLimit ? `<div class="c-textarea-limit" style="display:flex;justify-content:flex-end;font-size:0.75rem;color:#71717a;margin-top:0.35rem;"><span>${val.length} / ${charLimit} karakter</span></div>` : ''}
        </div>
      `;

      const ta = this.querySelector('textarea');
      ta.addEventListener('input', (e) => {
        this.setAttribute('value', e.target.value);
        if (valModel) Rikat.set(valModel, e.target.value);
        if (charLimit) {
          const lim = this.querySelector('.c-textarea-limit span');
          if (lim) lim.textContent = `${e.target.value.length} / ${charLimit} karakter`;
        }
        this.dispatchEvent(new CustomEvent('input', { detail: { value: e.target.value } }));
      });
    }
  }
  defElement('l-textarea', LTextarea);
  defElement('r-textarea', LTextarea);

  // LDropdown
  class LDropdown extends HTMLElement {
    static get observedAttributes() { return ['options', 'value', 'placeholder', 'disabled', 'icon', 'label']; }
    connectedCallback() { onReady(this, this.render); }
    attributeChangedCallback() { if (this._rendered) this.render(); }
    render() {
      this._rendered = true;
      const label = this.getAttribute('label');
      const ph = this.getAttribute('placeholder') || 'Pilih opsi...';
      const icon = this.getAttribute('icon');
      const disabled = this.hasAttribute('disabled');
      const valModel = this.getAttribute('[(value)]') || this.getAttribute('[(model)]');
      let curVal = this.getAttribute('value') || '';
      if (valModel && Rikat.get(valModel) !== undefined) curVal = String(Rikat.get(valModel));

      let opts = [];
      const rawOpts = this.getAttribute('options') || '[]';
      try {
        opts = JSON.parse(rawOpts);
      } catch(e) {
        opts = rawOpts.split(',').map(s => s.trim()).filter(Boolean);
      }

      const selectedOpt = opts.find(o => (typeof o === 'object' ? o.value === curVal : o === curVal));
      const displayLabel = selectedOpt ? (typeof selectedOpt === 'object' ? selectedOpt.label : selectedOpt) : ph;

      this.innerHTML = `
        <div class="c-dropdown-group" style="position:relative;width:100%;margin-bottom:0.5rem;">
          ${label ? `<label style="display:block;font-size:0.8rem;font-weight:600;color:#f4f4f5;margin-bottom:0.4rem;">${label}</label>` : ''}
          <div class="c-dropdown-trigger" style="display:flex;align-items:center;justify-content:space-between;gap:0.5rem;height:38px;padding:0 0.85rem;background:#161619;border:1px solid #27272a;border-radius:8px;cursor:${disabled ? 'not-allowed' : 'pointer'};opacity:${disabled ? '0.5' : '1'};user-select:none;transition:border-color 0.15s ease;">
            <div style="display:flex;align-items:center;gap:0.5rem;line-height:1.35;">
              ${icon ? `<l-icon name="${icon}" size="14" style="color:#71717a;"></l-icon>` : ''}
              <span style="font-size:0.875rem;color:${selectedOpt ? '#f4f4f5' : '#71717a'};font-weight:${selectedOpt ? '500' : '400'};">${displayLabel}</span>
            </div>
            <l-icon name="chevron-down" size="14" class="c-dropdown-chevron" style="color:#71717a;transition:transform 0.2s ease;"></l-icon>
          </div>
          <div class="c-dropdown-menu" style="display:none;position:absolute;top:calc(100% + 4px);left:0;right:0;background:#1a1a1e;border:1px solid #2d2d34;border-radius:8px;box-shadow:0 10px 25px -5px rgba(0,0,0,0.6);z-index:100;max-height:220px;overflow-y:auto;padding:4px;">
            ${opts.map(o => {
              const optVal = typeof o === 'object' ? o.value : o;
              const optLbl = typeof o === 'object' ? o.label : o;
              const optIcon = typeof o === 'object' ? o.icon : null;
              const isSel = (optVal === curVal);
              return `
                <div class="c-dropdown-item" data-value="${optVal}" style="display:flex;align-items:center;justify-content:space-between;gap:0.5rem;padding:0.55rem 0.75rem;border-radius:6px;cursor:pointer;font-size:0.85rem;color:${isSel ? '#ffffff' : '#d4d4d8'};background:${isSel ? 'rgba(255,45,32,0.15)' : 'transparent'};transition:background 0.1s ease;">
                  <div style="display:flex;align-items:center;gap:0.5rem;">
                    ${optIcon ? `<l-icon name="${optIcon}" size="14"></l-icon>` : ''}
                    <span>${optLbl}</span>
                  </div>
                  ${isSel ? '<l-icon name="check" size="14" style="color:#ff2d20;"></l-icon>' : ''}
                </div>
              `;
            }).join('')}
          </div>
        </div>
      `;

      if (!disabled) {
        const trigger = this.querySelector('.c-dropdown-trigger');
        const menu = this.querySelector('.c-dropdown-menu');
        const chevron = this.querySelector('.c-dropdown-chevron');

        const toggleMenu = (open) => {
          const isShow = open !== undefined ? open : (menu.style.display === 'none');
          menu.style.display = isShow ? 'block' : 'none';
          trigger.style.borderColor = isShow ? '#ff2d20' : '#27272a';
          chevron.style.transform = isShow ? 'rotate(180deg)' : 'rotate(0deg)';
        };

        trigger.addEventListener('click', (e) => {
          e.stopPropagation();
          toggleMenu();
        });

        menu.querySelectorAll('.c-dropdown-item').forEach(item => {
          item.addEventListener('mouseenter', () => { if (!item.style.background.includes('255,45,32')) item.style.background = '#27272e'; });
          item.addEventListener('mouseleave', () => { if (!item.style.background.includes('255,45,32')) item.style.background = 'transparent'; });
          item.addEventListener('click', (e) => {
            e.stopPropagation();
            const val = item.getAttribute('data-value');
            this.setAttribute('value', val);
            if (valModel) Rikat.set(valModel, val);
            this.dispatchEvent(new CustomEvent('change', { detail: { value: val } }));
            toggleMenu(false);
            this.render();
          });
        });

        const closeDoc = () => toggleMenu(false);
        document.addEventListener('click', closeDoc);
      }
    }
  }
  defElement('l-dropdown', LDropdown);
  defElement('r-dropdown', LDropdown);
  defElement('l-select', LDropdown);

  // LPagination
  class LPagination extends HTMLElement {
    static get observedAttributes() { return ['total-records', 'rows', 'page', 'show-report']; }
    connectedCallback() { onReady(this, this.render); }
    attributeChangedCallback() { if (this._rendered) this.render(); }
    render() {
      this._rendered = true;
      const total = parseInt(this.getAttribute('total-records') || this.getAttribute('total') || '100', 10);
      const rows = parseInt(this.getAttribute('rows') || '10', 10);
      let curPage = parseInt(this.getAttribute('page') || '1', 10);
      const totalPages = Math.max(1, Math.ceil(total / rows));
      curPage = Math.max(1, Math.min(totalPages, curPage));
      const showReport = this.hasAttribute('show-report');

      const startItem = (curPage - 1) * rows + 1;
      const endItem = Math.min(total, curPage * rows);

      let pageButtons = '';
      const maxButtons = 5;
      let startP = Math.max(1, curPage - Math.floor(maxButtons / 2));
      let endP = Math.min(totalPages, startP + maxButtons - 1);
      if (endP - startP + 1 < maxButtons) {
        startP = Math.max(1, endP - maxButtons + 1);
      }

      for (let p = startP; p <= endP; p++) {
        const isCur = (p === curPage);
        pageButtons += `
          <button type="button" class="c-page-btn" data-page="${p}" style="min-width:32px;height:32px;padding:0 6px;border-radius:6px;border:1px solid ${isCur ? '#ff2d20' : '#27272a'};background:${isCur ? '#ff2d20' : '#161619'};color:${isCur ? '#ffffff' : '#f4f4f5'};font-size:0.8rem;font-weight:600;cursor:pointer;display:inline-flex;align-items:center;justify-content:center;transition:all 0.15s ease;">
            ${p}
          </button>
        `;
      }

      this.innerHTML = `
        <div class="c-pagination-wrapper" style="display:flex;align-items:center;justify-content:space-between;flex-wrap:wrap;gap:1rem;width:100%;">
          ${showReport ? `<div class="c-pagination-report" style="font-size:0.8rem;color:#71717a;">Menampilkan <strong style="color:#f4f4f5;">${startItem}</strong> - <strong style="color:#f4f4f5;">${endItem}</strong> dari <strong style="color:#f4f4f5;">${total}</strong> data</div>` : '<div></div>'}
          <div class="c-pagination-controls" style="display:inline-flex;align-items:center;gap:0.35rem;">
            <button type="button" class="c-page-prev" ${curPage <= 1 ? 'disabled' : ''} style="width:32px;height:32px;border-radius:6px;border:1px solid #27272a;background:#161619;color:#f4f4f5;cursor:${curPage <= 1 ? 'not-allowed' : 'pointer'};opacity:${curPage <= 1 ? '0.4' : '1'};display:inline-flex;align-items:center;justify-content:center;">
              <l-icon name="chevron-left" size="14"></l-icon>
            </button>
            ${pageButtons}
            <button type="button" class="c-page-next" ${curPage >= totalPages ? 'disabled' : ''} style="width:32px;height:32px;border-radius:6px;border:1px solid #27272a;background:#161619;color:#f4f4f5;cursor:${curPage >= totalPages ? 'not-allowed' : 'pointer'};opacity:${curPage >= totalPages ? '0.4' : '1'};display:inline-flex;align-items:center;justify-content:center;">
              <l-icon name="chevron-right" size="14"></l-icon>
            </button>
          </div>
        </div>
      `;

      const setPage = (p) => {
        this.setAttribute('page', String(p));
        this.dispatchEvent(new CustomEvent('page-change', { detail: { page: p, rows } }));
        this.render();
      };

      this.querySelectorAll('.c-page-btn').forEach(b => {
        b.addEventListener('click', () => setPage(parseInt(b.getAttribute('data-page'), 10)));
      });
      const prevBtn = this.querySelector('.c-page-prev');
      if (prevBtn && curPage > 1) prevBtn.addEventListener('click', () => setPage(curPage - 1));
      const nextBtn = this.querySelector('.c-page-next');
      if (nextBtn && curPage < totalPages) nextBtn.addEventListener('click', () => setPage(curPage + 1));
    }
  }
  defElement('l-pagination', LPagination);
  defElement('r-pagination', LPagination);

  // LToast
  class LToast extends HTMLElement {
    static get observedAttributes() { return ['severity', 'summary', 'detail', 'closable']; }
    connectedCallback() { onReady(this, this.render); }
    render() {
      this._rendered = true;
      const severity = this.getAttribute('severity') || 'success';
      const summary = this.getAttribute('summary') || '';
      const detail = this.getAttribute('detail') || this.innerHTML.trim();
      const closable = this.hasAttribute('closable');
      const icon = severity === 'success' ? 'check-circle' : severity === 'danger' ? 'x-circle' : severity === 'warning' ? 'alert-circle' : 'info';
      const color = severity === 'success' ? '#10b981' : severity === 'danger' ? '#ef4444' : severity === 'warning' ? '#f59e0b' : '#3b82f6';

      this.innerHTML = `
        <div class="c-toast" style="display:flex;align-items:flex-start;gap:0.75rem;padding:0.85rem 1rem;background:#18181b;border:1px solid #27272a;border-left:4px solid ${color};border-radius:8px;box-shadow:0 10px 25px -5px rgba(0,0,0,0.5);min-width:300px;max-width:420px;position:relative;">
          <l-icon name="${icon}" size="18" style="color:${color};flex-shrink:0;margin-top:2px;"></l-icon>
          <div style="flex:1;">
            ${summary ? `<div style="font-size:0.875rem;font-weight:700;color:#ffffff;margin-bottom:0.2rem;">${summary}</div>` : ''}
            <div style="font-size:0.8rem;color:#a1a1aa;line-height:1.4;">${detail}</div>
          </div>
          ${closable ? `<button type="button" class="c-toast-close" style="all:unset;cursor:pointer;color:#71717a;display:flex;align-items:center;justify-content:center;padding:2px;" onmouseover="this.style.color='#f4f4f5'" onmouseout="this.style.color='#71717a'"><l-icon name="x" size="13"></l-icon></button>` : ''}
        </div>
      `;

      if (closable) {
        const btn = this.querySelector('.c-toast-close');
        if (btn) btn.addEventListener('click', () => {
          this.style.transition = 'opacity 0.2s, transform 0.2s';
          this.style.opacity = '0';
          this.style.transform = 'translateY(-6px)';
          setTimeout(() => this.remove(), 200);
        });
      }
    }
  }
  defElement('l-toast', LToast);
  defElement('r-toast', LToast);

  // LUpload
  class LUpload extends HTMLElement {
    static get observedAttributes() { return ['accept', 'multiple', 'label', 'hint']; }
    connectedCallback() { onReady(this, this.render); }
    render() {
      this._rendered = true;
      const label = this.getAttribute('label') || 'Tarik & Letakkan Berkas di Sini';
      const hint = this.getAttribute('hint') || 'Mendukung berkas PNG, JPG, PDF hingga 10MB';
      const accept = this.getAttribute('accept') || '*';
      const multiple = this.hasAttribute('multiple');

      this.innerHTML = `
        <div class="c-upload-wrapper" style="width:100%;">
          <div class="c-upload-dropzone" style="border:2px dashed #27272a;border-radius:10px;padding:2rem 1.5rem;text-align:center;background:#141416;cursor:pointer;transition:border-color 0.2s ease, background 0.2s ease;">
            <input type="file" style="display:none;" accept="${accept}" ${multiple ? 'multiple' : ''}>
            <div style="width:48px;height:48px;border-radius:50%;background:rgba(255,45,32,0.1);display:inline-flex;align-items:center;justify-content:center;color:#ff2d20;margin-bottom:0.75rem;">
              <l-icon name="upload-cloud" size="24"></l-icon>
            </div>
            <div style="font-size:0.95rem;font-weight:600;color:#f4f4f5;margin-bottom:0.25rem;">${label}</div>
            <div style="font-size:0.775rem;color:#71717a;margin-bottom:1rem;">${hint}</div>
            <l-button label="Pilih Berkas" severity="secondary" size="sm" icon="folder"></l-button>
          </div>
          <div class="c-upload-files" style="display:none;margin-top:0.75rem;display:flex;flex-direction:column;gap:0.5rem;"></div>
        </div>
      `;

      const dropzone = this.querySelector('.c-upload-dropzone');
      const fileInput = this.querySelector('input[type="file"]');
      const fileList = this.querySelector('.c-upload-files');

      dropzone.addEventListener('click', () => fileInput.click());
      dropzone.addEventListener('dragover', (e) => {
        e.preventDefault();
        dropzone.style.borderColor = '#ff2d20';
        dropzone.style.background = 'rgba(255,45,32,0.05)';
      });
      dropzone.addEventListener('dragleave', () => {
        dropzone.style.borderColor = '#27272a';
        dropzone.style.background = '#141416';
      });
      dropzone.addEventListener('drop', (e) => {
        e.preventDefault();
        dropzone.style.borderColor = '#27272a';
        dropzone.style.background = '#141416';
        if (e.dataTransfer.files.length) handleFiles(e.dataTransfer.files);
      });

      fileInput.addEventListener('change', () => {
        if (fileInput.files.length) handleFiles(fileInput.files);
      });

      const handleFiles = (files) => {
        fileList.innerHTML = '';
        Array.from(files).forEach(f => {
          const item = document.createElement('div');
          item.style.cssText = 'display:flex;align-items:center;justify-content:space-between;padding:0.5rem 0.75rem;background:#1a1a1e;border:1px solid #27272a;border-radius:6px;font-size:0.8rem;color:#f4f4f5;';
          const sz = (f.size / 1024).toFixed(1) + ' KB';
          item.innerHTML = `
            <div style="display:flex;align-items:center;gap:0.5rem;overflow:hidden;">
              <l-icon name="file" size="14" style="color:#ff2d20;"></l-icon>
              <span style="line-height:1.35;">${f.name}</span>
              <span style="color:#71717a;font-size:0.75rem;">(${sz})</span>
            </div>
            <span style="color:#10b981;font-size:0.75rem;font-weight:600;"><l-icon name="check" size="13"></l-icon> Siap</span>
          `;
          fileList.appendChild(item);
        });
        this.dispatchEvent(new CustomEvent('select', { detail: { files: Array.from(files) } }));
      };
    }
  }
  defElement('l-upload', LUpload);
  defElement('r-upload', LUpload);
  defElement('l-file-upload', LUpload);

  // LButtonGroup
  class LButtonGroup extends HTMLElement {
    connectedCallback() { onReady(this, this.render); }
    render() {
      this.style.display = 'inline-flex';
      this.style.borderRadius = '8px';
      this.style.overflow = 'hidden';
      this.style.border = '1px solid #27272a';
      Array.from(this.children).forEach((child, idx, arr) => {
        if (child.tagName.toLowerCase().includes('button')) {
          const btn = child.querySelector('button') || child;
          btn.style.borderRadius = '0';
          btn.style.border = 'none';
          if (idx < arr.length - 1) {
            btn.style.borderRight = '1px solid #27272a';
          }
        }
      });
    }
  }
  defElement('l-button-group', LButtonGroup);
  defElement('r-button-group', LButtonGroup);

  // Global Rikat.toast helper
  Rikat.toast = {
    show: function(opts) {
      let container = document.getElementById('rikat-toast-container');
      if (!container) {
        container = document.createElement('div');
        container.id = 'rikat-toast-container';
        container.style.cssText = 'position:fixed;top:1.5rem;right:1.5rem;z-index:9999;display:flex;flex-direction:column;gap:0.75rem;pointer-events:none;';
        document.body.appendChild(container);
      }
      const t = document.createElement('l-toast');
      t.setAttribute('severity', opts.severity || 'info');
      if (opts.summary) t.setAttribute('summary', opts.summary);
      if (opts.detail) t.setAttribute('detail', opts.detail);
      t.setAttribute('closable', '');
      t.style.pointerEvents = 'auto';
      t.style.animation = 'rikatSlideIn 0.25s cubic-bezier(0.16, 1, 0.3, 1)';
      container.appendChild(t);
      if (opts.life !== 0) {
        setTimeout(() => {
          t.style.transition = 'opacity 0.3s, transform 0.3s';
          t.style.opacity = '0';
          t.style.transform = 'translateY(-10px)';
          setTimeout(() => t.remove(), 300);
        }, opts.life || 3000);
      }
    }
  };


  // Auto-initialize on DOMContentLoaded
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => Rikat.init());
  } else {
    Rikat.init();
  }
})();

