// VoicePilot-Eino Web Client
class VoicePilotClient {
    constructor() {
        this.baseURL = window.location.origin;
        this.sessionId = this.generateSessionId();
        this.mediaRecorder = null;
        this.audioChunks = [];
        this.isRecording = false;

        this.init();
    }

    init() {
        this.setupEventListeners();
        this.updateSessionDisplay();
        this.checkServerStatus();
    }

    generateSessionId() {
        return 'session_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
    }

    setupEventListeners() {
        const recordBtn = document.getElementById('recordBtn');
        const uploadBtn = document.getElementById('uploadBtn');
        const fileInput = document.getElementById('fileInput');
        const sendTextBtn = document.getElementById('sendTextBtn');
        const textInput = document.getElementById('textInput');

        // Recording button - press and hold
        recordBtn.addEventListener('mousedown', () => this.startRecording());
        recordBtn.addEventListener('mouseup', () => this.stopRecording());
        recordBtn.addEventListener('touchstart', (e) => {
            e.preventDefault();
            this.startRecording();
        });
        recordBtn.addEventListener('touchend', (e) => {
            e.preventDefault();
            this.stopRecording();
        });

        // Upload button
        uploadBtn.addEventListener('click', () => fileInput.click());
        fileInput.addEventListener('change', (e) => this.handleFileUpload(e));

        // Text input
        sendTextBtn.addEventListener('click', () => this.sendTextMessage());
        textInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.sendTextMessage();
            }
        });
    }

    async checkServerStatus() {
        try {
            const response = await fetch(`${this.baseURL}/api/health`);
            if (response.ok) {
                this.updateStatus('connected', '已连接');
            } else {
                this.updateStatus('error', '服务器错误');
            }
        } catch (error) {
            this.updateStatus('error', '无法连接到服务器');
            console.error('Server status check failed:', error);
        }
    }

    updateStatus(status, text) {
        const indicator = document.getElementById('statusIndicator');
        const statusText = document.getElementById('statusText');

        indicator.className = 'status-indicator status-' + status;
        statusText.textContent = text;
    }

    updateSessionDisplay() {
        document.getElementById('sessionId').textContent = this.sessionId;
    }

    showLoading(message = '处理中...') {
        const overlay = document.getElementById('loadingOverlay');
        const text = document.getElementById('loadingText');
        text.textContent = message;
        overlay.style.display = 'flex';
    }

    hideLoading() {
        document.getElementById('loadingOverlay').style.display = 'none';
    }

    async startRecording() {
        if (this.isRecording) return;

        try {
            const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
            this.mediaRecorder = new MediaRecorder(stream);
            this.audioChunks = [];

            this.mediaRecorder.addEventListener('dataavailable', event => {
                this.audioChunks.push(event.data);
            });

            this.mediaRecorder.addEventListener('stop', () => {
                const audioBlob = new Blob(this.audioChunks, { type: 'audio/wav' });
                this.sendVoiceRequest(audioBlob);

                // Stop all tracks
                stream.getTracks().forEach(track => track.stop());
            });

            this.mediaRecorder.start();
            this.isRecording = true;
            this.updateStatus('recording', '正在录音...');

            const recordBtn = document.getElementById('recordBtn');
            recordBtn.classList.add('recording');
            recordBtn.querySelector('.btn-text').textContent = '松开发送';

        } catch (error) {
            console.error('Failed to start recording:', error);
            this.updateStatus('error', '麦克风权限被拒绝');
            alert('无法访问麦克风，请检查浏览器权限设置。');
        }
    }

    stopRecording() {
        if (!this.isRecording) return;

        this.mediaRecorder.stop();
        this.isRecording = false;
        this.updateStatus('processing', '处理中...');

        const recordBtn = document.getElementById('recordBtn');
        recordBtn.classList.remove('recording');
        recordBtn.querySelector('.btn-text').textContent = '按住说话';
    }

    async sendVoiceRequest(audioBlob) {
        this.showLoading('正在识别语音...');

        try {
            const formData = new FormData();
            formData.append('audio', audioBlob, 'recording.wav');
            formData.append('session_id', this.sessionId);

            const response = await fetch(`${this.baseURL}/api/voice`, {
                method: 'POST',
                body: formData
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            this.handleVoiceResponse(data);

        } catch (error) {
            console.error('Voice request failed:', error);
            this.updateStatus('error', '请求失败');
            this.addConversationItem('system', '处理失败: ' + error.message);
        } finally {
            this.hideLoading();
        }
    }

    async handleFileUpload(event) {
        const file = event.target.files[0];
        if (!file) return;

        // Validate file type
        if (!file.type.startsWith('audio/')) {
            alert('请选择音频文件（WAV或MP3格式）');
            return;
        }

        // Validate file size (10MB max)
        if (file.size > 10 * 1024 * 1024) {
            alert('文件大小不能超过10MB');
            return;
        }

        this.showLoading('正在上传音频文件...');

        try {
            const formData = new FormData();
            formData.append('audio', file);
            formData.append('session_id', this.sessionId);

            const response = await fetch(`${this.baseURL}/api/voice`, {
                method: 'POST',
                body: formData
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            this.handleVoiceResponse(data);

        } catch (error) {
            console.error('File upload failed:', error);
            this.updateStatus('error', '上传失败');
            this.addConversationItem('system', '上传失败: ' + error.message);
        } finally {
            this.hideLoading();
            event.target.value = ''; // Reset file input
        }
    }

    async sendTextMessage() {
        const textInput = document.getElementById('textInput');
        const text = textInput.value.trim();

        if (!text) {
            alert('请输入文字内容');
            return;
        }

        this.showLoading('正在处理...');

        // Add user message to conversation
        this.addConversationItem('user', text);

        // Clear input
        textInput.value = '';

        try {
            const response = await fetch(`${this.baseURL}/api/text`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    text: text,
                    session_id: this.sessionId
                })
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            this.handleVoiceResponse(data);

        } catch (error) {
            console.error('Text request failed:', error);
            this.updateStatus('error', '请求失败');
            this.addConversationItem('system', '处理失败: ' + error.message);
        } finally {
            this.hideLoading();
        }
    }

    handleVoiceResponse(data) {
        if (!data.success) {
            this.updateStatus('error', '处理失败');
            this.addConversationItem('system', data.error || '未知错误');
            return;
        }

        this.updateStatus('connected', '已连接');

        // Add conversation items
        this.addConversationItem('assistant', data.text);

        // Play audio response if available
        if (data.audio_url) {
            this.playAudioResponse(data.audio_url);
        }

        // Update session ID if changed
        if (data.session_id && data.session_id !== this.sessionId) {
            this.sessionId = data.session_id;
            this.updateSessionDisplay();
        }
    }

    addConversationItem(role, message) {
        const conversationList = document.getElementById('conversationList');

        // Remove empty state if it exists
        const emptyState = conversationList.querySelector('.empty-state');
        if (emptyState) {
            emptyState.remove();
        }

        const item = document.createElement('div');
        item.className = 'conversation-item conversation-' + role;

        const avatar = document.createElement('div');
        avatar.className = 'avatar';
        avatar.textContent = role === 'user' ? '👤' : (role === 'assistant' ? '🤖' : '⚙️');

        const content = document.createElement('div');
        content.className = 'content';

        const roleName = document.createElement('div');
        roleName.className = 'role';
        roleName.textContent = role === 'user' ? '你' : (role === 'assistant' ? '助手' : '系统');

        const messageText = document.createElement('div');
        messageText.className = 'message';
        messageText.textContent = message;

        const timestamp = document.createElement('div');
        timestamp.className = 'timestamp';
        timestamp.textContent = new Date().toLocaleTimeString('zh-CN');

        content.appendChild(roleName);
        content.appendChild(messageText);
        content.appendChild(timestamp);

        item.appendChild(avatar);
        item.appendChild(content);

        conversationList.appendChild(item);

        // Scroll to bottom
        conversationList.scrollTop = conversationList.scrollHeight;
    }

    playAudioResponse(audioUrl) {
        const audioPanel = document.getElementById('audioPanel');
        const audioPlayer = document.getElementById('audioPlayer');

        // Construct full URL if relative
        const fullUrl = audioUrl.startsWith('http') ? audioUrl : this.baseURL + audioUrl;

        audioPlayer.src = fullUrl;
        audioPanel.style.display = 'block';

        // Auto play
        audioPlayer.play().catch(error => {
            console.error('Failed to play audio:', error);
        });
    }
}

// Initialize the client when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    window.voicePilot = new VoicePilotClient();
});
