document.addEventListener('alpine:init', () => {
    Alpine.data('shortenerApp', () => ({
        mode: 'create', 
        
        url: '',
        alias: '',
        showAlias: false,
        shortenResult: null,

        inspectCode: '',
        inspectResult: null,

        loading: false,
        error: null,

        async submitShorten() {
            this.loading = true;
            this.error = null;
            this.shortenResult = null;

            try {
                const response = await fetch('/shorten', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ url: this.url, alias: this.alias })
                });
                const data = await response.json();
                
                if (!response.ok) throw new Error(data.error || 'Erro ao encurtar');

                this.shortenResult = data;
                this.url = '';
                this.alias = '';
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
            this.inspectResult = null;

            try {
                const response = await fetch('/inspect', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ code: this.inspectCode })
                });
                const data = await response.json();

                if (!response.ok) throw new Error(data.error || 'Link não encontrado');

                this.inspectResult = data;
            } catch (err) {
                this.error = err.message;
            } finally {
                this.loading = false;
            }
        },

        copyLink() {
            if (!this.shortenResult) return;
            navigator.clipboard.writeText(this.shortenResult.short_url);
            alert("Link copiado!");
        }
    }));
});