export const formatSize = (bytes) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

export const getExt = (f) => f.split('.').pop().toLowerCase();

export const isImage = (f) => ['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg'].includes(getExt(f));
export const isVideo = (f) => ['mp4', 'webm', 'ogg'].includes(getExt(f));
export const isAudio = (f) => ['mp3', 'wav', 'm4a'].includes(getExt(f));
export const isText = (f) => ['txt', 'md', 'json', 'js', 'css', 'go', 'py', 'sql'].includes(getExt(f));

export const getFileIcon = (f) => {
    if (isImage(f)) return 'fas fa-image';
    if (isVideo(f)) return 'fas fa-video';
    if (isAudio(f)) return 'fas fa-music';
    const ext = getExt(f);
    if (ext === 'pdf') return 'fas fa-file-pdf';
    if (isText(f)) return 'fas fa-file-code';
    return 'fas fa-file';
};

export const getFileUrl = (f) => `/api/storage?key=${encodeURIComponent(f)}`;
