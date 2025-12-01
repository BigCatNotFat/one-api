import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import { Container } from 'semantic-ui-react';
import App from './App';
import Header from './components/Header';
import Footer from './components/Footer';
import 'semantic-ui-css/semantic.min.css';
import './index.css';
import { UserProvider } from './context/User';
import { ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import { StatusProvider } from './context/Status';
import { useLocation } from 'react-router-dom';
import './i18n';

const Layout = () => {
  const location = useLocation();
  const isQueryPage = location.pathname === '/query';

  if (isQueryPage) {
    return (
      <>
        <App />
        <ToastContainer />
      </>
    );
  }

  return (
    <>
      <Header />
      <Container className={'main-content'}>
        <App />
      </Container>
      <ToastContainer />
      <Footer />
    </>
  );
};

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
  <React.StrictMode>
    <StatusProvider>
      <UserProvider>
        <BrowserRouter>
          <Layout />
        </BrowserRouter>
      </UserProvider>
    </StatusProvider>
  </React.StrictMode>
);
