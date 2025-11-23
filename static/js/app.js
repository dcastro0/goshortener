document.addEventListener('alpine:init', () => {
    Alpine.data('shortenerApp', () => ({
        mode: 'create',
        
        url: '',
        alias: '',
        password: '', 
        showAlias: false,
        shortenResult: null,

        // Inspect
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

                if (data.error) {
                    throw new Error(data.error);
                }

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
});