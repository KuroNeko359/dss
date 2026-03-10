export const fetchList = async () => {
    const response = await fetch('/api/list');
    if (!response.ok) throw new Error('获取列表失败');
    return await response.json();
};

export const uploadFile = async (file) => {
    const formData = new FormData();
    formData.append('file', file);
    const response = await fetch('/api/upload', { method: 'POST', body: formData });
    if (!response.ok) throw new Error(await response.text());
    return true;
};

export const deleteFile = async (key) => {
    const response = await fetch(`/api/storage?key=${encodeURIComponent(key)}`, { method: 'DELETE' });
    if (!response.ok) throw new Error('删除失败');
    return true;
};

export const getFileContent = async (url) => {
    const response = await fetch(url);
    if (!response.ok) throw new Error('获取内容失败');
    return await response.text();
};
