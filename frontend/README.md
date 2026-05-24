# Multi-Carrier Shipping Platform - Frontend

A modern, responsive React frontend for managing multi-carrier shipping operations. Built with React 18, Vite, and styled for an intuitive user experience.

## 🚀 Features

- **Dashboard** - Overview of shipments, pending items, and key metrics
- **Shipment Management** - View, create, update, and track shipments
- **List & Filter** - Browse shipments with status filtering
- **Shipment Details** - View complete shipment information and tracking history
- **Create Shipment** - Intuitive form to create new shipments with carrier selection
- **API Testing** - Built-in API testing console for developers
- **Settings** - Configure API base URL and authentication token
- **Responsive Design** - Works great on desktop and mobile devices

## 📦 Installation

### Prerequisites
- Node.js 16+ 
- npm or yarn

### Quick Start

```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

The app will be available at `http://localhost:5173` (or the URL printed by Vite).

## 🔧 Configuration

### Environment Variables

Create a `.env` file in the `frontend/` directory:

```env
VITE_API_URL=http://localhost:8080
```

### API Configuration

1. Navigate to Settings (⚙️)
2. Configure API Base URL (defaults to `http://localhost:8080`)
3. Set Authorization Token (defaults to `Bearer test-token`)

Or set via environment variables before running Vite.

## 📱 Pages & Components

### Dashboard
Quick overview with statistics:
- Total shipments count
- Pending shipments count
- Return requests count
- Pending invoices count

### Shipments List
Browse all shipments with features:
- Status filtering (All, Pending, Processing, Delivered, Cancelled)
- Quick view details button
- Responsive data table
- View creation dates

### Shipment Details
Complete shipment information:
- Basic shipment metadata
- Sender information
- Receiver information
- Package details (weight, dimensions, cost)
- Tracking history timeline

### Create Shipment
User-friendly form to create new shipments:
- Sender information (name, address, email)
- Receiver information (name, address, email)
- Package details (weight, dimensions)
- Carrier selection (DHL, FedEx, UPS, USPS)
- Service type selection (Standard, Express, Overnight, Economy)

### API Testing Console
Developer tools for testing endpoints:
- Pre-configured forms for all API endpoints
- Manual request building
- JSON response viewing
- Useful for debugging and development

### Settings
Configuration page:
- API Base URL configuration
- Authentication token management
- Application information

## 🎯 Navigation

Use the sidebar to navigate between sections:
- 📊 Dashboard - Overview
- 📋 Shipments - List all shipments
- ➕ Create Shipment - Create new shipment
- 🧪 API Test - Test endpoints
- ⚙️ Settings - Configure app

## 🔗 API Integration

The frontend integrates with the API Gateway at `/api` with these main endpoint groups:

- **Shipments** - CRUD operations and status updates
- **Carriers** - Carrier registration and rate queries
- **Tracking** - Real-time tracking information
- **Rates** - Rate comparison between carriers
- **Labels** - Label generation and management
- **Addresses** - Address validation and location queries
- **Billing** - Invoice creation and payment processing
- **Returns** - Return request management

## 🎨 Styling

The application uses a custom CSS styling system with:
- Flexbox and CSS Grid layouts
- Responsive design (mobile-first)
- Modern color scheme with blue primary
- Smooth transitions and hover effects
- Accessible form controls

### Color Scheme
- Primary Blue: `#3b82f6`
- Gray Neutral: `#6b7280`
- Success Green: `#10b981`
- Warning Amber: `#f59e0b`
- Error Red: `#ef4444`

## 📁 Project Structure

```
frontend/
├── src/
│   ├── components/
│   │   ├── ApiForm.jsx           # Reusable API testing form
│   │   ├── Dashboard.jsx         # Dashboard overview
│   │   ├── ShipmentList.jsx      # Shipment list with filters
│   │   ├── ShipmentDetail.jsx    # Detailed shipment view
│   │   ├── CreateShipment.jsx    # Create shipment form
│   │   └── Settings.jsx          # Settings page
│   ├── services/
│   │   └── api.js               # API service layer
│   ├── App.jsx                  # Main app component
│   ├── main.jsx                 # Entry point
│   └── styles.css               # Global styles
├── index.html
├── package.json
├── vite.config.js
└── README.md
```

## 🚀 Deployment

### Build for Production
```bash
npm run build
```

Output files are in the `dist/` directory, ready to be served by any static hosting service.

### Docker Example
```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build
EXPOSE 5173
CMD ["npm", "run", "preview"]
```

## 🐛 Troubleshooting

### API Connection Issues
- Verify the API Gateway is running on `http://localhost:8080`
- Check the Authorization token in Settings
- Open browser DevTools to check error messages

### Shipments Not Loading
- Ensure backend services are healthy
- Check database migrations have been applied
- Verify authentication token has proper permissions

### Port Already in Use
```bash
# Use a different port
npm run dev -- --port 5174
```

## 📚 Backend Documentation

- [API Guide](/docs/API-GUIDE.md) - Backend API documentation
- [Architecture](/docs/ARCHITECTURE.md) - System architecture
- [Services](/docs/SERVICES.md) - Service descriptions

## 👨‍💻 Development

### Technology Stack
- **React 18.2.0** - UI framework
- **Vite 5.0.0** - Build tool and dev server
- **CSS3** - Styling (no CSS framework dependencies)
- **Fetch API** - HTTP requests

### Code Style
- Modern ES6+ JavaScript
- Functional React components with hooks
- Single responsibility principle
- Clear separation of concerns

### Adding New Components

1. Create component file in `src/components/YourComponent.jsx`
2. Import in `App.jsx`
3. Add case to `renderContent()` switch statement
4. Add navigation button in sidebar (if needed)

Example:
```jsx
// src/components/Orders.jsx
export default function Orders() {
  return <div className="orders"><h1>Orders</h1></div>
}
```

Update `App.jsx`:
```jsx
import Orders from './components/Orders'

// In renderContent():
case 'orders':
  return <Orders />

// In navigation:
<button className="nav-item" onClick={() => setView('orders')}>
  📦 Orders
</button>
```

## 🔐 Security Notes

- Tokens are stored in component state (not recommended for production)
- No token encryption or secure storage
- For production, use proper auth patterns (OAuth, JWT with secure storage)
- Don't expose secrets in environment variables sent to browser
- Always use HTTPS in production

## 📈 Future Enhancements

Potential features to add:
- User authentication with proper session management
- Real-time notifications for shipment updates
- Advanced filtering and search
- Export/report generation
- Bulk shipment operations
- Integration with carrier webhooks
- Analytics and metrics dashboard
- Multi-user support with roles
- Audit logging

## 📝 License

Part of the Multi-Carrier Shipping Platform

