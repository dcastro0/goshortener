document.addEventListener('alpine:init', () => {
    Alpine.data('shortenerApp', () => ({
        mode: 'create',
        url: '',
        alias: '',
        password: '',
        showAlias: false,
        shortenResult: null,
        inspectCode: '',
        inspectPassword: '',
        inspectResult: null,
        loading: false,
        error: null,

        init() {
            const params = new URLSearchParams(window.location.search);
            const inspectHash = params.get('inspect');
            if (inspectHash) {
                this.mode = 'inspect';
                this.inspectCode = inspectHash;
                this.submitInspect();
                window.history.replaceState({}, document.title, "/");
            }
        },

        async submitShorten() {
            this.loading = true;
            this.error = null;
            this.shortenResult = null;
            try {
                const response = await fetch('/shorten', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ 
                        url: this.url, 
                        alias: this.alias,
                        password: this.password 
                    })
                });
                const data = await response.json();
                if (!response.ok) throw new Error(data.error || 'Erro ao encurtar');
                this.shortenResult = data;
                this.url = '';
                this.alias = '';
                this.password = '';
                this.showAlias = false;
            } catch (err) {
                this.error = err.message;
            } finally {
                this.loading = false;
            }
        },

        async submitInspect() {
            this.loading = true;
            this.error = null;
            try {
                const response = await fetch('/inspect', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ 
                        code: this.inspectCode,
                        password: this.inspectPassword 
                    })
                });
                const data = await response.json();
                if (!response.ok) throw new Error(data.error || 'Link não encontrado');
                if (data.error) throw new Error(data.error);
                this.inspectResult = data;
                this.inspectPassword = '';
            } catch (err) {
                this.error = err.message;
            } finally {
                this.loading = false;
            }
        },

        copyLink() {
            if (!this.shortenResult) return;
            navigator.clipboard.writeText(this.shortenResult.short_url);
        }
    }));

    Alpine.data('statsApp', (chartData = []) => ({
        isEditModalOpen: false,
        editingId: null,
        editForm: {
            url: '',
            alias: ''
        },
        loading: false,

        init() {
            this.renderChart(chartData);
        },

        renderChart(data) {
            const ctx = document.getElementById('clicksChart');
            if (!ctx) return;
            const topLinks = data.slice(0, 10);
            new Chart(ctx.getContext('2d'), {
                type: 'bar',
                data: {
                    labels: topLinks.map(l => l.hash),
                    datasets: [{
                        label: 'Cliques',
                        data: topLinks.map(l => l.clicks),
                        backgroundColor: 'rgba(59, 130, 246, 0.5)',
                        borderColor: 'rgba(59, 130, 246, 1)',
                        borderWidth: 1,
                        borderRadius: 4
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    scales: {
                        y: { beginAtZero: true, grid: { color: 'rgba(255, 255, 255, 0.1)' }, ticks: { color: '#94a3b8' } },
                        x: { grid: { display: false }, ticks: { color: '#94a3b8' } }
                    },
                    plugins: { legend: { display: false } }
                }
            });
        },

        openEditModal(link) {
            this.editingId = link.id;
            this.editForm.url = link.url;
            this.editForm.alias = link.hash;
            this.isEditModalOpen = true;
        },

        async submitEdit() {
            this.loading = true;
            try {
                const res = await fetch('/link/' + this.editingId, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(this.editForm)
                });
                const data = await res.json();

                if (res.ok) {
                    alert('Link atualizado! A página será recarregada.');
                    window.location.reload(); 
                } else {
                    alert('Erro: ' + data.error);
                }
            } catch (e) {
                console.error(e);
                alert('Erro de conexão');
            } finally {
                this.loading = false;
            }
        },

        async deleteItem(btn, id, type) {
            const text = type === 'link' ? 'este link' : 'esta mensagem';
            if (!confirm('Tem certeza que deseja excluir ' + text + ' permanentemente?')) return;

            btn.disabled = true;
            btn.classList.add('opacity-50', 'cursor-not-allowed');

            try {
                const res = await fetch('/' + type + '/' + id, { method: 'DELETE' });
                if (res.ok) {
                    const row = btn.closest('tr');
                    row.style.transition = 'all 0.3s ease';
                    row.style.opacity = '0';
                    setTimeout(() => row.remove(), 300);
                } else {
                    alert('Erro ao excluir');
                    btn.disabled = false;
                    btn.classList.remove('opacity-50', 'cursor-not-allowed');
                }
            } catch (e) {
                console.error(e);
                alert('Erro de conexão');
                btn.disabled = false;
                btn.classList.remove('opacity-50', 'cursor-not-allowed');
            }
        }
    }));
});