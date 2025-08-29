window.fileManager = function() {
    return {
        //files: [
        //    {{ range .Files }}
        //            { id: '{{ .ID }}', name: '{{ .Filename }}', sizeBytes: {{ .Size }}, expiresAt: '{{ .Expiresat }}' },
        //        {{ end }}
        //    // Example files with proper format
        //    //{ id: 'f1', name: 'document.pdf', sizeBytes: 2621440, expiresAt: '2025-06-20 20:55:06.163887 +0300 EEST' },
        //    //{ id: 'f2', name: 'image.jpg', sizeBytes: 1258291, expiresAt: '2025-06-15 09:15:00.000000 +0300 EEST' },
        //    //{ id: 'f3', name: 'spreadsheet.xlsx', sizeBytes: 876544, expiresAt: '2025-06-25 18:45:00.000000 +0300 EEST' },
        //    //{ id: 'f4', name: 'presentation.pptx', sizeBytes: 4299161, expiresAt: '2025-06-14 12:00:00.000000 +0300 EEST' }
        //],
        //copied: false,
        //showUploadModal: false,
        //showDeleteModal: false,
        //selectedFiles: [],
        //fileToDelete: null,
        //progressWidth: 0,
        //uploadProgress: 0,
        //uploadStatus: 'idle', // idle, uploading, complete, error
        //showSpinner: true,

        //// Server can set these values in bytes
        //usedBytes: {{ .Space }},//7301365760, // Example: 6.8 GB in bytes
        //totalBytes: {{ .MAX }}, //10737418240, // Example: 10 GB in bytes
        //calcBytes: 0,
        //filename: "",

        init() {
            // Start progress bar animation after 100ms
            setTimeout(() => {
                this.progressWidth = this.storagePercentage;
            }, 100);
        },

        get storagePercentage() {
            return Math.round((this.usedBytes / this.totalBytes) * 100);
        },

        handleFileSelection(event) {
            this.selectedFiles = Array.from(event.target.files);
            for ( const file of this.selectedFiles ) {
                this.calcBytes += file.size
            }
        },

        cancelUpload() {
            this.resetUploadModal();
        },
        //I need work

        continueUpload(uploadID) {
            if (this.selectedFiles.length === 0) return;

            this.uploadStatus = 'uploading';
            this.uploadProgress = 0;

            // Create EventSource for progress updates
            const eventSource = new EventSource(`/status?id=${uploadID}`);

            eventSource.onopen = () => {
                this.showSpinner = true;
                setTimeout(() => {
                    this.performUpload(uploadID);
                }, 2000);
            };

            eventSource.onmessage = (event) => {
                this.showSpinner = false;
                //console.log(event.data)
                const data = JSON.parse(event.data);

                if (data.percentage !== undefined) {
                    this.uploadProgress = Math.round(data.percentage);
                }
                if (data.total_bytes === -1) {
                    this.uploadProgress = Math.round((data.bytes/this.calcBytes) * 100)
                }

                if (data.message === 'Upload complete') {
                    this.uploadStatus = 'complete';
                    setTimeout(() => {
                        this.resetUploadModal();
                        this.showSpinner = true;
                        //// Refresh file list or add new files to the list
                        location.reload(); // Simple refresh, or implement dynamic update
                        eventSource.close();
                    }, 1500);
                } else if (data.status === 'error') {
                    this.uploadStatus = 'error';
                    eventSource.close();
                }
            };

            eventSource.onerror = (error) => {
                console.error('EventSource failed:', error);
                this.uploadStatus = 'error';
                eventSource.close();
            };

            // Start the upload AFTER EventSource is ready (onopen event)
        },

        async performUpload(uploadID) {
            const formData = new FormData();

            // Add the custom filename, the user sent
            formData.append(`filename`, this.filename);

            // Add all selected files to FormData
            this.selectedFiles.forEach((file) => {
                formData.append(`file`, file);
            });

            try {
                const response = await fetch(`/upload?id=${uploadID}`, {
                    method: 'POST',
                    body: formData
                });
                if (!response.ok) {
                    throw new Error('Upload failed');
                }
            } catch (e) {
                console.error('Upload error:', e);
                this.uploadStatus = 'error';
            }
        },

        resetUploadModal() {
            this.showUploadModal = false;
            this.selectedFiles = [];
            this.uploadStatus = 'idle';
            this.uploadProgress = 0;
        },

        confirmDelete(file) {
            this.fileToDelete = file;
            this.showDeleteModal = true;
        },

        async deleteFile() {
            if (this.fileToDelete) {
                const delData = new FormData();
                delData.append(`file`, this.fileToDelete.id)
                try {
                    const response = await fetch(`/delete/`, {
                        method: 'POST',
                        body: delData,
                    })
                    if (!response.ok) {
                        throw new Error('Delete failed:');
                    }
                    location.reload(); // Simple refresh, or implement dynamic update
                } catch (e) {
                    console.error(e);
                }
                this.files = this.files.filter(f => f.id !== this.fileToDelete.id);
                this.showDeleteModal = false;
                this.fileToDelete = null;
            }
        },

        downloadFile(file, server, user) {
            //alert('Downloading: ' + file.name);
            //Possibly add an ssl thing variable
            const url = `http://${server}/download/${user}/${file.id}`;
            window.location.href = url;
        },

        copyDownloadUrl(file, server, user) {
            const url = `http://${server}/download/${user}/${file.id}`;
            navigator.clipboard.writeText(url).then(() => {
                //alert('Download URL copied to clipboard');
            });
        },

        formatBytes(bytes) {
            if (bytes === 0) return '0 Bytes';
            const k = 1024;
            const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
        },

        formatFileSize(bytes) {
            return this.formatBytes(bytes);
        },

        // Parse server time format: "2025-06-20 20:55:06.163887 +0300 EEST"
        parseServerTime(timeStr) {
            // Remove timezone abbreviation and parse
            const cleanedTime = timeStr.replace(/\s+[A-Z]{3,4}$/, '');

            // Handle the timezone offset format
            // Convert "2025-06-20 20:55:06.163887 +0300" to ISO format
            const parts = cleanedTime.split(' ');
            if (parts.length >= 3) {
                const datePart = parts[0];
                const timePart = parts[1];
                const offsetPart = parts[2];

                // Reformat to ISO string
                const isoString = `${datePart}T${timePart}${offsetPart.slice(0,3)}:${offsetPart.slice(3)}`;
                return new Date(isoString);
            }

            // Fallback: try direct parsing
            return new Date(timeStr);
        },

        getTimeLeft(expiresAtStr) {
            const now = new Date();
            const expiresAt = this.parseServerTime(expiresAtStr);
            const diffMs = expiresAt - now;

            if (diffMs <= 0) {
                return 'Expired';
            }

            const days = Math.floor(diffMs / (1000 * 60 * 60 * 24));
            const hours = Math.floor((diffMs % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));

            if (days > 0) {
                return `${days}d ${hours}h`;
            } else if (hours > 0) {
                return `${hours}h`;
            } else {
                const minutes = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60));
                return `${minutes}m`;
            }
        },

        getTimeLeftClass(expiresAtStr) {
            const now = new Date();
            const expiresAt = this.parseServerTime(expiresAtStr);
            const diffMs = expiresAt - now;
            const hours = diffMs / (1000 * 60 * 60);

            if (diffMs <= 0) {
                return 'text-red-600 dark:text-red-400 font-medium';
            } else if (hours <= 24) {
                return 'text-orange-600 dark:text-orange-400 font-medium';
            } else if (hours <= 72) {
                return 'text-yellow-600 dark:text-yellow-400';
            } else {
                return 'text-slate-500 dark:text-slate-400';
            }
        },

        logout(){
            window.location.href = '/logout';
        }
    }
}
