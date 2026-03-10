import { createApp, ref, computed, onMounted } from 'https://unpkg.com/vue@3/dist/vue.esm-browser.js';
import * as api from './modules/api.js';
import * as utils from './modules/utils.js';

const app = createApp({
    setup() {
        const keyList = ref([]);
        const searchQuery = ref('');
        const currentTab = ref('all');
        const viewMode = ref('grid');
        const showUploadModal = ref(false);
        const selectedFile = ref(null);
        const isUploading = ref(false);
        const isDragging = ref(false);
        const isRefreshing = ref(false);
        const message = ref('');
        const messageType = ref('success');
        const previewingFile = ref(null);
        const textContent = ref('');

        const currentTabName = computed(() => {
            const names = { 'all': '全部文件', 'images': '图片', 'media': '视频 / 音频', 'docs': '文档' };
            return names[currentTab.value];
        });

        const filteredFiles = computed(() => {
            let files = keyList.value;
            if (searchQuery.value) {
                files = files.filter(f => f.toLowerCase().includes(searchQuery.value.toLowerCase()));
            }
            if (currentTab.value === 'images') {
                files = files.filter(f => utils.isImage(f));
            } else if (currentTab.value === 'media') {
                files = files.filter(f => utils.isVideo(f) || utils.isAudio(f));
            } else if (currentTab.value === 'docs') {
                files = files.filter(f => utils.isText(f) || f.endsWith('.pdf'));
            }
            return files;
        });

        const showMessage = (msg, type = 'success') => {
            message.value = msg;
            messageType.value = type;
            setTimeout(() => message.value = '', 3000);
        };

        const fetchList = async () => {
            isRefreshing.value = true;
            try {
                const data = await api.fetchList();
                keyList.value = data;
            } catch (err) {
                showMessage(err.message, 'error');
            } finally {
                setTimeout(() => isRefreshing.value = false, 500);
            }
        };

        const handleFileChange = (e) => {
            selectedFile.value = e.target.files[0];
        };

        const handleDrop = (e) => {
            isDragging.value = false;
            selectedFile.value = e.dataTransfer.files[0];
        };

        const handleUpload = async () => {
            if (!selectedFile.value) return;
            isUploading.value = true;
            try {
                await api.uploadFile(selectedFile.value);
                showMessage('文件已安全存入 DSS');
                selectedFile.value = null;
                showUploadModal.value = false;
                fetchList();
            } catch (err) {
                showMessage('上传失败: ' + err.message, 'error');
            } finally {
                isUploading.value = false;
            }
        };

        const handleDelete = async (key) => {
            if (!confirm(`确定要彻底删除 ${key} 吗？`)) return;
            try {
                await api.deleteFile(key);
                showMessage('文件已从存储中移除');
                if (previewingFile.value === key) previewingFile.value = null;
                fetchList();
            } catch (err) {
                showMessage(err.message, 'error');
            }
        };

        const previewFile = async (file) => {
            previewingFile.value = file;
            if (utils.isText(file)) {
                textContent.value = '正在加载内容...';
                try {
                    textContent.value = await api.getFileContent(utils.getFileUrl(file));
                } catch (e) {
                    textContent.value = '无法加载文本内容';
                }
            }
        };

        onMounted(fetchList);

        return {
            keyList, searchQuery, currentTab, viewMode, showUploadModal, selectedFile,
            isUploading, isDragging, isRefreshing, message, messageType,
            previewingFile, textContent, currentTabName, filteredFiles,
            fetchList, handleFileChange, handleDrop, handleUpload, handleDelete,
            previewFile, 
            // Utils
            formatSize: utils.formatSize,
            getFileIcon: utils.getFileIcon,
            getFileUrl: utils.getFileUrl,
            getExt: utils.getExt,
            isImage: utils.isImage,
            isVideo: utils.isVideo,
            isAudio: utils.isAudio,
            isText: utils.isText
        };
    }
});

app.mount('#app');
