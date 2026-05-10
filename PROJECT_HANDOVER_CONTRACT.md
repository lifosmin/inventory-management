# PROJECT HANDOVER CONTRACT
## Lot-Based Inventory Accounting System

**Date:** May 7, 2026  
**Project Name:** Lot-Based Inventory Accounting System (Full-Stack)  
**Developer:** [Your Name]  
**Client:** [Client Name]

---

## 1. PROJECT OVERVIEW

This contract outlines the terms and conditions for the handover of the **Lot-Based Inventory Accounting System**, a comprehensive full-stack inventory management solution consisting of:

- **Backend API**: Built with Go and PostgreSQL for robust data management
- **Frontend Application**: Built with Next.js, React, and TypeScript for modern user interface

The system is designed to track inventory by discrete lots (batches) with full traceability of cost, quantity, age, and movement history, providing both API access and a complete web-based user interface.

---

## 2. SERVICE MODEL

**Important:** The project is **NOT transferred to client ownership**. The Developer provides a **service-based model** where:

- The Developer retains ownership and responsibility for the system infrastructure
- The Client receives access to the system and its services
- The Developer manages deployment, hosting, and system maintenance
- The System operates in a Software-as-a-Service (SaaS) arrangement

---

## 3. PROJECT FEATURES & API ENDPOINTS

### 3.1 Authentication & Authorization
- **User Login** - JWT token generation with HMAC-SHA256
- **Token Refresh** - Refresh token rotation for session management
- **User Logout** - Session termination
- **Role-Based Access Control (RBAC)** - User role management and permission enforcement

### 3.2 User Management
- **Create User** - Add new system users with assigned roles
- **Read User** - Retrieve user profiles and information
- **Update User** - Modify user details and role assignments
- **Delete User** - Remove user accounts from the system
- **List Users** - View all users with pagination

### 3.3 Product Management
- **Create Product** - Register new products in the system
- **Read Product** - Retrieve product details and specifications
- **Update Product** - Modify product information
- **Delete Product** - Remove products from the system
- **List Products** - View all products with category filtering
- **Product Categories** - Category management for product organization

### 3.4 Warehouse Management
- **Create Warehouse** - Register new warehouse locations
- **Read Warehouse** - Retrieve warehouse information
- **Update Warehouse** - Modify warehouse details
- **Delete Warehouse** - Remove warehouse records
- **List Warehouses** - View all warehouses
- **Warehouse Locations** - Manage storage locations within warehouses

### 3.5 Lot Management (Core Feature)
- **Create Lot** - Register new inventory lots/batches with cost tracking
- **Read Lot** - Retrieve lot details and full traceability information
- **Update Lot** - Modify lot information (limited fields)
- **Delete Lot** - Remove lot records
- **List Lots** - View all lots with advanced filtering
- **Lot Aging** - Track lot age and aging information
- **Lot Costing** - Calculate and track lot cost basis
- **FIFO Strategy** - First-In-First-Out allocation strategy for warehouse operations

### 3.6 Inventory Management
- **Stock Movements** - Record inventory transfers between locations/warehouses
- **Inventory Adjustments** - Reconcile inventory discrepancies
- **Stock Allocation** - Allocate lots to fulfill orders
- **Movement History** - Full audit trail of all inventory movements
- **Real-time Stock Levels** - View current stock quantities and valuations

### 3.7 Sales & Shipment Management
- **Create Sale Order** - Generate sales orders with lot allocation
- **Manage Sales** - Update and track sales orders
- **Shipment Tracking** - Track shipments from warehouse to delivery
- **Payment Tracking** - Monitor payment status and reconciliation

### 3.8 Reporting & Analytics
- **Inventory Valuation Report** - Calculate total inventory value using lot costing
- **Lot Aging Report** - Identify slow-moving and aging inventory
- **Stock Movement Report** - Historical analysis of inventory movements
- **Sales Report** - Sales performance and order tracking
- **Warehouse Utilization Report** - Monitor warehouse capacity and usage

### 3.9 System Features
- **Request Rate Limiting** - Prevent API abuse with rate limiting
- **CORS Policy** - Secure cross-origin requests for frontend integration
- **Request Tracing** - Request ID injection for debugging and monitoring
- **Structured Logging** - Comprehensive system logging for monitoring
- **Database Migrations** - Version-controlled schema management
- **Docker Deployment** - Containerized deployment ready

---

## 4. FRONTEND APPLICATION FEATURES

### 4.1 User Interface & Navigation
- **Responsive Design** - Mobile and desktop compatible interface
- **Modern UI Framework** - Built with Next.js 16, React 19, and TypeScript
- **Tailwind CSS Styling** - Professional, consistent visual design
- **Navigation System** - Intuitive menu structure for all features

### 4.2 Dashboard
- **Overview Dashboard** - Real-time system status and key metrics
- **Quick Access** - Fast navigation to main functions
- **Status Indicators** - Visual feedback on system health

### 4.3 Sales Management Interface
- **Sales Order Creation** - Create and manage sales orders with lot allocation
- **Payment Status Tracking** - Monitor payment progress (Unpaid/DP/Fully Paid)
- **Sales History** - Complete sales transaction history with sorting and filtering
- **Real-time Updates** - Live data synchronization with backend
- **Search & Filter** - Advanced search capabilities for sales records

### 4.4 Inventory Restocking Interface
- **Lot Management** - Create and manage inventory lots with automatic lot numbering
- **Product Selection** - Searchable product selection for restocking
- **Warehouse Assignment** - Assign lots to specific warehouse locations
- **Cost Tracking** - Track purchase costs and pricing per lot
- **Payment Status** - Monitor restocking payment status
- **Bulk Operations** - Efficient handling of multiple restocking operations

### 4.5 Reports & Analytics Interface
- **Inventory Valuation Reports** - Real-time inventory value calculations
- **Margin Analysis** - Visual margin percentage indicators with color coding
- **Collection Tracking** - Payment collection progress with progress bars
- **Export Functionality** - Export reports in various formats
- **Date Range Filtering** - Custom date range selection for reports
- **Financial Metrics** - Revenue, costs, and profit margin calculations

### 4.6 Settings & Administration Interface
- **Product Management** - Create, edit, and manage product catalog
- **Warehouse Management** - Configure warehouses and FIFO strategies
- **User Profile** - View and manage user account information
- **System Configuration** - Administrative settings and preferences
- **Data Management** - Bulk operations for products and warehouses

### 4.7 Authentication Interface
- **Secure Login** - JWT-based authentication with secure token management
- **Session Management** - Automatic session handling and refresh
- **Access Control** - Role-based UI access and feature restrictions

### 4.8 Technical Features
- **Real-time Data** - Live updates without page refresh
- **Error Handling** - User-friendly error messages and validation
- **Loading States** - Professional loading indicators and feedback
- **Form Validation** - Client-side and server-side validation
- **Responsive Tables** - Sortable, filterable data tables
- **Search Functionality** - Global search across all modules

---

## 4. DEVELOPER RESPONSIBILITIES POST-HANDOVER

After the contract and application handover, the Developer will provide:

### 4.1 Small Bug Fixes (Included in Monthly Fee)
- Bug fixes for minor issues (non-critical, non-blocking functionality)
- Issues that do not significantly impact system operation
- Estimated resolution time: < 24 hours per ticket
- Response time: Within 5 business days

### 4.2 Excluded Services (Additional Pricing Required)
The following services require additional pricing negotiation:

- **Major Feature Updates** - New API endpoints or significant system enhancements
- **Major Bug Fixes** - Critical bugs affecting core functionality, data integrity, or system availability
- **Performance Optimization** - System scaling, database optimization, caching implementation
- **Security Patches** - Security vulnerabilities or security-related updates
- **Third-party Integration** - Integration with external systems or services
- **Architecture Changes** - Significant system redesign or refactoring
- **Consulting & Custom Development** - Custom business logic or specialized requirements

---

## 5. PAYMENT TERMS & PRICING

### 5.1 Initial Project Fee
| Item | Amount |
|------|--------|
| **Total Website Development** | **Rp 5,000,000** |

Payment is due upon signing of this contract.

### 5.2 Monthly Recurring Fees (Begins upon handover)
| Service | Amount | Frequency |
|---------|--------|-----------|
| **Deployment & Hosting Fee** | Rp 100,000 | Monthly |
| **Free Support** | Rp 0 | For 1 month from handover* |
| **Paid Support** | Rp 50,000 | After 1st month |

**Free Support Includes:**
- Small bug fixes (as defined in Section 4.1)
- Technical support and consultation
- System monitoring and maintenance
- Database backups and security updates

*Free support period starts when both the contract is signed AND the application is handed over to the Client

**Additional charges will apply for services listed in Section 4.2

### 5.3 Payment Schedule
- Initial project fee: Due upon contract signature
- Monthly hosting fee: Due on the 1st of each month, valid for 30 days
- Overdue payments: 2% monthly interest on unpaid balance

---

## 6. SUPPORT & MAINTENANCE

### 6.1 First Month (Free Support - From Handover Date)
- **Availability:** Monday - Friday, 9 AM - 5 PM [Timezone]
- **Response Time:** Within 24 business hours for inquiries
- **Scope:** Small bug fixes, minor adjustments, basic support
- **Database Maintenance:** Automatic backups, monitoring, security updates

### 6.2 Included in Monthly Fee (Rp 100,000)
- Infrastructure hosting and deployment
- Database maintenance and backups
- SSL/TLS certificate management
- Basic system monitoring
- Uptime maintenance (99% SLA target)
- Security patches and OS updates

---

## 7. SERVICE LEVEL AGREEMENT (SLA)

| Metric | Target |
|--------|--------|
| **System Uptime** | 99% per month |
| **Response Time** | Within 24 business hours |
| **Bug Fix Acknowledgment** | Within 4 business hours |
| **Database Backup Frequency** | Daily |
| **Emergency Support** | Available upon request (additional fee applies) |

---

## 8. TECHNICAL SPECIFICATIONS

**Technology Stack:**
- **Backend:** Go (net/http + chi router)
- **Frontend:** Next.js 16, React 19, TypeScript, Tailwind CSS
- **Database:** PostgreSQL 16+
- **Authentication:** JWT with HMAC-SHA256
- **Deployment:** Docker & Docker Compose
- **API Protocol:** REST (JSON)
- **UI Framework:** Responsive web application

**Current Status:**
- ✅ Backend API fully functional
- ✅ Frontend Application complete and integrated
- ✅ Database schema and migrations complete
- ✅ Authentication and authorization implemented
- ✅ Core inventory management features operational
- ✅ User interface for all major functions implemented

---

## 9. TERMS & CONDITIONS

### 9.1 Ownership & Intellectual Property
- The Client receives a non-exclusive license to use the system
- The Developer retains ownership of the source code and intellectual property
- The Client may not redistribute, resell, or sublicense the system without written permission

### 9.2 Data Protection & Privacy
- The Developer is responsible for data security and backups
- The Client is responsible for compliance with local data protection regulations
- Regular security audits will be conducted (frequency TBD)

### 9.3 Termination
- Either party may terminate this agreement with 30 days written notice
- Upon termination, the Client loses access to the system
- The Client's data will be retained for 30 days post-termination for migration purposes

### 9.4 Liability
- The Developer is not liable for data loss caused by Client negligence
- The Developer is not liable for third-party service outages
- Maximum liability is limited to the fees paid in the preceding 3 months

### 9.5 Modification Rights
- The Developer reserves the right to modify system infrastructure for security or performance
- The Developer will provide 7 days notice for major infrastructure changes
- The Developer may perform maintenance with advance notice (preferably off-hours)

---

## 10. REVISION & UPDATES POLICY

### Major Feature Requests
Any new features not listed in Section 3 require:
- Detailed requirements documentation
- Time and cost estimation
- Separate contract amendment
- Pricing based on complexity (start from Rp 500,000 - Rp 5,000,000+ per feature)

### Major Bug Fixes
Issues requiring significant code changes or architectural modifications require:
- Issue severity assessment
- Impact analysis
- Cost estimation
- Separate quote and contract amendment

### Pricing Model
- **Simple fixes (< 4 hours):** Rp 200,000 - Rp 500,000
- **Medium complexity (4-16 hours):** Rp 500,000 - Rp 2,000,000
- **Complex projects (> 16 hours):** Custom pricing based on scope

---

## 11. ACCEPTANCE & SIGNATURES

By signing below, both parties agree to the terms and conditions outlined in this contract.

### Developer
```
Name: _____________________________
Signature: _________________________
Date: _____________________________
Contact: __________________________
```

### Client
```
Name: _____________________________
Company: ___________________________
Signature: _________________________
Date: _____________________________
Contact: __________________________
```

---

## 12. APPENDICES

### Appendix A: API Documentation
A complete API documentation will be provided separately with endpoint details, request/response examples, and authentication requirements.

### Appendix B: Setup & Deployment Guide
A comprehensive setup guide for local development and deployment procedures.

### Appendix C: Data Export Policy
Upon contract termination, the Client may request their data in JSON/CSV format within 30 days.

---

**Document Version:** 1.0  
**Last Updated:** May 7, 2026  
**Contract Status:** [Ready for Signature / Draft / Signed]

---

*This is a template. Please customize names, dates, timezone, payment terms, and any other specific details before presenting to the client.*
