const $ = (id) => document.getElementById(id);
let me = null, cur = null, ws = null, seen = new Set(), timer = null;

async function api(path, method = 'GET', body) {
  const r = await fetch('/api/' + path, {
    method,
    headers: { 'Content-Type': 'application/json' },
    body: body && JSON.stringify(body),
  });
  const j = await r.json().catch(() => ({}));
  if (!r.ok) throw new Error(j.error || 'Ошибка запроса');
  return j;
}

const fmt = (iso) => new Date(iso).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });

// ---------- вход и регистрация ----------
async function submitAuth(register) {
  $('err').textContent = '';
  const username = $('u').value, password = $('p').value;
  try {
    if (register) {
      await api('register', 'POST', { username, display_name: $('d').value || username, password });
    }
    await api('login', 'POST', { username, password });
    await start();
  } catch (e) {
    $('err').textContent = e.message;
  }
}
$('login').onclick = () => submitAuth(false);
$('reg').onclick = () => submitAuth(true);
$('p').onkeydown = (e) => { if (e.key === 'Enter') submitAuth(false); };

$('out').onclick = async () => {
  await api('logout', 'POST');
  me = null; cur = null;
  if (ws) ws.close();
  $('app').hidden = true; $('auth').hidden = false;
};

async function start() {
  me = await api('me');
  $('auth').hidden = true; $('app').hidden = false;
  $('me').textContent = me.display_name;
  await loadChats();
  connect();
}

// ---------- WebSocket ----------
function connect() {
  clearTimeout(timer);
  ws = new WebSocket((location.protocol === 'https:' ? 'wss://' : 'ws://') + location.host + '/ws');
  ws.onmessage = (e) => {
    const d = JSON.parse(e.data);
    if (d.type !== 'message') return;
    if (cur && cur.id === d.chat_id) addMsg(d.message);
    loadChats();
  };
  ws.onclose = () => { if (me) timer = setTimeout(connect, 2000); };
}

// ---------- список чатов и поиск ----------
async function loadChats() {
  if ($('q').value.trim()) return; // сейчас показан поиск
  const chats = await api('chats');
  const list = $('list');
  list.replaceChildren();
  if (!chats.length) {
    list.append(el('div', 'empty', 'Чатов пока нет. Найдите пользователя в поиске выше.'));
    return;
  }
  for (const c of chats) {
    const it = el('div', 'item' + (cur && cur.id === c.id ? ' on' : ''));
    it.append(el('b', '', c.name), el('small', '', c.last || 'Нет сообщений'));
    it.onclick = () => openChat({ id: c.id, name: c.name });
    list.append(it);
  }
}

function el(tag, cls, text) {
  const n = document.createElement(tag);
  if (cls) n.className = cls;
  if (text !== undefined) n.textContent = text;
  return n;
}

let searchTimer;
$('q').oninput = () => {
  clearTimeout(searchTimer);
  searchTimer = setTimeout(async () => {
    const q = $('q').value.trim();
    if (!q) return loadChats();
    const users = await api('users?q=' + encodeURIComponent(q));
    const list = $('list');
    list.replaceChildren();
    if (!users.length) list.append(el('div', 'empty', 'Никого не нашли'));
    for (const u of users) {
      const it = el('div', 'item');
      it.append(el('b', '', u.display_name), el('small', '', '@' + u.username));
      it.onclick = async () => {
        const { id } = await api('chats', 'POST', { user_id: u.id });
        $('q').value = '';
        await openChat({ id, name: u.display_name });
        loadChats();
      };
      list.append(it);
    }
  }, 250);
};

// ---------- переписка ----------
async function openChat(chat) {
  cur = chat; seen = new Set();
  $('title').textContent = chat.name;
  $('app').classList.add('chat');
  $('text').disabled = $('send').disabled = false;
  $('msgs').replaceChildren();
  const msgs = await api(`chats/${chat.id}/messages`);
  msgs.forEach(addMsg);
  loadChats();
  $('text').focus();
}
$('back').onclick = () => { $('app').classList.remove('chat'); cur = null; };

function addMsg(m) {
  if (seen.has(m.id)) return;
  seen.add(m.id);
  const box = $('msgs');
  const d = el('div', 'msg' + (m.sender_id === me.id ? ' mine' : ''), m.content);
  d.append(el('time', '', fmt(m.created_at)));
  box.append(d);
  box.scrollTop = box.scrollHeight;
}

async function sendMsg() {
  const content = $('text').value.trim();
  if (!content || !cur) return;
  $('text').value = '';
  try {
    addMsg(await api(`chats/${cur.id}/messages`, 'POST', { content }));
    loadChats();
  } catch (e) {
    $('text').value = content;
    alert(e.message);
  }
}
$('send').onclick = sendMsg;
$('text').onkeydown = (e) => { if (e.key === 'Enter') sendMsg(); };

// автоматический вход, если cookie ещё действует
api('me').then(start).catch(() => {});
