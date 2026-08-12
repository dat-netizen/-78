import React, { useState, useEffect } from 'react';

const API_BASE = window.location.origin.includes('-3000.')
  ? window.location.origin.replace('-3000.', '-8080.')
  : `${window.location.protocol}//${window.location.hostname}:8080`;

function App() {
  const [tableNo, setTableNo] = useState('1');
  const [menuItems, setMenuItems] = useState([]);
  const [cart, setCart] = useState([]);
  const [orders, setOrders] = useState([]);
  const [activeTab, setActiveTab] = useState('menu'); // menu, confirm, history, checkout
  const [message, setMessage] = useState('');

  const fetchOptions = (options = {}) => {
    return {
      ...options,
      credentials: 'include',
      headers: {
        ...options.headers,
        'X-Table-No': String(tableNo),
      },
    };
  };

  useEffect(() => {
    fetchMenu();
    fetchOrders();
  }, [tableNo]);

  const fetchMenu = async () => {
    try {
      const res = await fetch(`${API_BASE}/api/menu`, fetchOptions());
      if (res.ok) {
        const data = await res.json();
        setMenuItems(data || []);
      }
    } catch (e) {
      console.error('メニュー取得エラー:', e);
    }
  };

  const fetchOrders = async () => {
    try {
      const res = await fetch(`${API_BASE}/api/orders`, fetchOptions());
      if (res.ok) {
        const data = await res.json();
        setOrders(data || []);
      }
    } catch (e) {
      console.error('注文履歴取得エラー:', e);
    }
  };

  const addToCart = (item) => {
    setCart((prevCart) => {
      const existing = prevCart.find((ci) => ci.id === item.id);
      if (existing) {
        return prevCart.map((ci) =>
          ci.id === item.id ? { ...ci, quantity: ci.quantity + 1 } : ci
        );
      }
      return [...prevCart, { ...item, quantity: 1 }];
    });
    showMessage(`${item.name} をカートに追加しました`);
  };

  const updateCartQuantity = (id, delta) => {
    setCart((prevCart) => {
      return prevCart
        .map((item) => {
          if (item.id === id) {
            const newQty = item.quantity + delta;
            return newQty > 0 ? { ...item, quantity: newQty } : null;
          }
          return item;
        })
        .filter(Boolean);
    });
  };

  const showMessage = (msg) => {
    setMessage(msg);
    setTimeout(() => setMessage(''), 3000);
  };

  const handleOrderSubmit = async () => {
    if (cart.length === 0) return;

    const payload = {
      items: cart.map((item) => ({
        menu_item_id: item.id,
        quantity: item.quantity,
      })),
    };

    try {
      const res = await fetch(`${API_BASE}/api/orders`, fetchOptions({
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      }));

      if (res.ok) {
        setCart([]);
        await fetchOrders();
        setActiveTab('history');
        showMessage('ご注文を承りました！');
      } else {
        alert('注文送信に失敗しました');
      }
    } catch (e) {
      console.error(e);
      alert('通信エラーが発生しました');
    }
  };

  const handleCheckout = async () => {
    if (!window.confirm('お会計を確定しますか？')) return;

    try {
      const res = await fetch(`${API_BASE}/api/orders/checkout`, fetchOptions({
        method: 'POST',
      }));

      if (res.ok) {
        await fetchOrders();
        setActiveTab('menu');
        showMessage('お会計が完了しました。ご来店ありがとうございました！');
      }
    } catch (e) {
      console.error(e);
      alert('会計処理に失敗しました');
    }
  };

  const cartTotal = cart.reduce((sum, item) => sum + item.price * item.quantity, 0);
  const pendingOrders = orders.filter((o) => o.status === 'pending');
  const unpaidTotal = pendingOrders.reduce((sum, o) => sum + o.total_amount, 0);

  return (
    <div className="app-container">
      <header className="navbar">
        <div className="nav-content">
          <h1 className="logo" onClick={() => setActiveTab('menu')}>
            🇻🇳 ベトナム食堂
          </h1>
          <div className="table-selector">
            <label htmlFor="table-select">卓番号: </label>
            <input
              id="table-select"
              type="text"
              value={tableNo}
              onChange={(e) => setTableNo(e.target.value)}
            />
          </div>
        </div>
      </header>

      {message && <div className="toast-message">{message}</div>}

      <nav className="sub-nav">
        <button
          className={activeTab === 'menu' ? 'active' : ''}
          onClick={() => setActiveTab('menu')}
        >
          メニュー選択
        </button>
        <button
          className={activeTab === 'confirm' ? 'active' : ''}
          onClick={() => setActiveTab('confirm')}
        >
          注文確認 ({cart.reduce((s, i) => s + i.quantity, 0)})
        </button>
        <button
          className={activeTab === 'history' ? 'active' : ''}
          onClick={() => setActiveTab('history')}
        >
          注文履歴
        </button>
        <button
          className={activeTab === 'checkout' ? 'active' : ''}
          onClick={() => setActiveTab('checkout')}
        >
          お会計
        </button>
      </nav>

      <main className="main-layout">
        {activeTab === 'menu' && (
          <section className="menu-grid">
            {menuItems.map((item) => (
              <div key={item.id} className="menu-card card">
                <img src={item.image_url} alt={item.name} className="menu-img" />
                <div className="menu-info">
                  <div className="menu-header">
                    <h3>{item.name}</h3>
                    <span className="vietnamese">{item.vietnamese}</span>
                  </div>
                  <p className="description">{item.description}</p>
                  <div className="menu-bottom">
                    <span className="price">¥{item.price.toLocaleString()}</span>
                    <button className="btn-primary" onClick={() => addToCart(item)}>
                      追加
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </section>
        )}

        {activeTab === 'confirm' && (
          <section className="confirm-section card">
            <h2>現在のカート内容</h2>
            {cart.length === 0 ? (
              <p className="empty-message">カートに商品が入っていません。</p>
            ) : (
              <>
                <div className="cart-list">
                  {cart.map((item) => (
                    <div key={item.id} className="cart-item">
                      <div className="cart-item-info">
                        <h4>{item.name}</h4>
                        <span className="price">¥{item.price.toLocaleString()}</span>
                      </div>
                      <div className="qty-controls">
                        <button onClick={() => updateCartQuantity(item.id, -1)}>-</button>
                        <span>{item.quantity}</span>
                        <button onClick={() => updateCartQuantity(item.id, 1)}>+</button>
                      </div>
                    </div>
                  ))}
                </div>
                <div className="summary-box">
                  <h3>合計: ¥{cartTotal.toLocaleString()}</h3>
                  <button className="btn-submit" onClick={handleOrderSubmit}>
                    注文を確定する
                  </button>
                </div>
              </>
            )}
          </section>
        )}

        {activeTab === 'history' && (
          <section className="history-section">
            <h2>卓番号 {tableNo} の注文履歴</h2>
            {orders.length === 0 ? (
              <p className="empty-message">注文履歴がありません。</p>
            ) : (
              orders.map((order) => (
                <div key={order.id} className="order-card card">
                  <div className="order-header">
                    <span>注文ID: #{order.id}</span>
                    <span className={`status-badge ${order.status}`}>
                      {order.status === 'pending' ? '調理中・未会計' : '会計済み'}
                    </span>
                  </div>
                  <div className="order-items">
                    {order.items.map((item) => (
                      <div key={item.id} className="order-item-row">
                        <span>{item.name} x {item.quantity}</span>
                        <span>¥{(item.price * item.quantity).toLocaleString()}</span>
                      </div>
                    ))}
                  </div>
                  <div className="order-footer">
                    <span>日時: {new Date(order.created_at).toLocaleTimeString()}</span>
                    <strong>小計: ¥{order.total_amount.toLocaleString()}</strong>
                  </div>
                </div>
              ))
            )}
          </section>
        )}

        {activeTab === 'checkout' && (
          <section className="checkout-section card">
            <h2>お会計（卓番号: {tableNo}）</h2>
            {unpaidTotal === 0 ? (
              <p className="empty-message">現在お支払いが必要な未精算の注文はありません。</p>
            ) : (
              <>
                <div className="unpaid-summary">
                  <h3>未清算の注文一覧</h3>
                  {pendingOrders.map((o) => (
                    <div key={o.id} className="unpaid-order-row">
                      <span>注文ID: #{o.id}</span>
                      <span>¥{o.total_amount.toLocaleString()}</span>
                    </div>
                  ))}
                  <div className="total-due">
                    <h3>お支払い合計: ¥{unpaidTotal.toLocaleString()}</h3>
                  </div>
                </div>
                <button className="btn-checkout" onClick={handleCheckout}>
                  お会計を完了する
                </button>
              </>
            )}
          </section>
        )}
      </main>
    </div>
  );
}

export default App;