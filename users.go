package backenduser1

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	pasproj "github.com/e-dumas-sukasari/webpasetobackend"
	"github.com/gofiber/fiber/v2"
)

func SubmitReport(MongoEnv, dbname, colname, publickey string, r *http.Request) string {
    resp := new(pasproj.Credential)
    req := new(pasproj.Report)
    conn := pasproj.MongoCreateConnection(MongoEnv, dbname)
    tokenlogin := r.Header.Get("Login")

    if tokenlogin == "" {
        resp.Status = false
        resp.Message = "Header Login Not Found"
    } else {
        checkHR := IsHR(tokenlogin, os.Getenv(publickey))
        if !checkHR {
            resp.Status = false
            resp.Message = "Anda tidak memiliki izin untuk mengirim laporan"
        } else {
            // Parse JSON request body
            decoder := json.NewDecoder(r.Body)
            if err := decoder.Decode(&req); err != nil {
                resp.Status = false
                resp.Message = "Error decoding JSON request: " + err.Error()
                return pasproj.ReturnStringStruct(resp)
            }

            // Perform your logic to save the report data to the database using conn
            // Example: pasproj.InsertReportData(conn, colname, req)
            if err := pasproj.InsertReportData(conn, colname, req); err != nil {
                resp.Status = false
                resp.Message = "Error inserting report data: " + err.Error()
                return pasproj.ReturnStringStruct(resp)
            }

            resp.Status = true
            resp.Message = "Laporan berhasil dikirim"
        }
    }
    return pasproj.ReturnStringStruct(resp)
}



// reg User
func Register(Mongoenv, dbname string, r *http.Request) string {
	resp := new(pasproj.Credential)
	userdata := new(pasproj.User)
	resp.Status = false
	conn := pasproj.MongoCreateConnection(Mongoenv, dbname)
	err := json.NewDecoder(r.Body).Decode(&userdata)
	if err != nil {
		resp.Message = "error parsing application/json: " + err.Error()
	} else {
		resp.Status = true
		hash, err := pasproj.HashPass(userdata.Password)
		if err != nil {
			resp.Message = "Gagal Hash Password" + err.Error()
		}
		pasproj.InsertUserdata(conn, userdata.Username, userdata.Role, hash)
		resp.Message = "Berhasil Input data"
	}
	response := pasproj.ReturnStringStruct(resp)
	return response

}

// log User
func Login(Privatekey, MongoEnv, dbname, Colname string, r *http.Request) string {
	var resp pasproj.Credential
	mconn := pasproj.MongoCreateConnection(MongoEnv, dbname)
	var datauser pasproj.User
	err := json.NewDecoder(r.Body).Decode(&datauser)
	if err != nil {
		resp.Message = "error parsing application/json: " + err.Error()
	} else {
		if pasproj.PasswordValidator(mconn, Colname, datauser) {
			datarole := pasproj.GetOneUser(mconn, "user", pasproj.User{Username: datauser.Username})
			tokenstring, err := pasproj.EncodeWithRole(datarole.Role, datauser.Username, os.Getenv(Privatekey))
			if err != nil {
				resp.Message = "Gagal Encode Token : " + err.Error()
			} else {
				resp.Status = true
				resp.Message = "Selamat Datang"
				resp.Token = tokenstring
			}
		} else {
			resp.Message = "Password Salah"
		}
	}
	return pasproj.ReturnStringStruct(resp)
}

// Get Data User
func GetDataUserForAdmin(PublicKey, MongoEnv, dbname, colname string, r *http.Request) string {
	req := new(pasproj.ResponseDataUser)
	conn := pasproj.MongoCreateConnection(MongoEnv, dbname)
	tokenlogin := r.Header.Get("Login")
	if tokenlogin == "" {
		req.Status = false
		req.Message = "Header Login Not Found"
	} else {
		cekadmin := IsAdmin(tokenlogin, PublicKey)
		if cekadmin != true {
			req.Status = false
			req.Message = "IHHH Kamu bukan admin"
		}
		checktoken, err := pasproj.DecodeGetUser(os.Getenv(PublicKey), tokenlogin)
		if err != nil {
			req.Status = false
			req.Message = "tidak ada data username : " + tokenlogin
		}
		compared := pasproj.CompareUsername(conn, colname, checktoken)
		if compared != true {
			req.Status = false
			req.Message = "Data User tidak ada"
		} else {
			datauser := pasproj.GetAllUser(conn, colname)
			req.Status = true
			req.Message = "data User berhasil diambil"
			req.Data = datauser
		}
	}
	return pasproj.ReturnStringStruct(req)
}

// Reset Password
func ResetPassword(MongoEnv, publickey, dbname, colname string, r *http.Request) string {
	resp := new(Cred)
	req := new(pasproj.User)
	conn := pasproj.MongoCreateConnection(MongoEnv, dbname)
	tokenlogin := r.Header.Get("Login")
	if tokenlogin == "" {
		resp.Status = fiber.StatusBadRequest
		resp.Message = "Token login tidak ada"
	} else {
		checkadmin := IsAdmin(tokenlogin, os.Getenv(publickey))
		if !checkadmin {
			resp.Status = fiber.StatusInternalServerError
			resp.Message = "kamu bukan admin"
		} else {
			UpdatePassword(conn, pasproj.User{
				Username: req.Username,
				Password: req.Password,
			})
			resp.Status = fiber.StatusOK
			resp.Message = "Berhasil reset password"
		}
	}
	return pasproj.ReturnStringStruct(resp)
}

// Delete User
func DeleteUserforAdmin(Mongoenv, publickey, dbname, colname string, r *http.Request) string {
	resp := new(Cred)
	req := new(ReqUsers)
	conn := pasproj.MongoCreateConnection(Mongoenv, dbname)
	tokenlogin := r.Header.Get("Login")
	if tokenlogin == "" {
		resp.Status = fiber.StatusBadRequest
		resp.Message = "Token login tidak ada"
	} else {
		checkadmin := IsAdmin(tokenlogin, os.Getenv(publickey))
		if !checkadmin {
			resp.Status = fiber.StatusInternalServerError
			resp.Message = "kamu bukan admin"
		} else {
			_, err := DeleteUser(conn, colname, req.Username)
			if err != nil {
				resp.Status = fiber.StatusBadRequest
				resp.Message = "gagal hapus data"
			}
			resp.Status = fiber.StatusOK
			resp.Message = "data berhasil dihapus"
		}
	}
	return pasproj.ReturnStringStruct(resp)
}

// Insert data
func InsertEmployee(MongoEnv, dbname, colname, publickey string, r *http.Request) string {
	resp := new(pasproj.Credential)
	req := new(Employee)
	conn := pasproj.MongoCreateConnection(MongoEnv, dbname)
	tokenlogin := r.Header.Get("Login")
	if tokenlogin == "" {
		resp.Status = false
		resp.Message = "Header Login Not Found"
	} else {
		checkadmin := IsAdmin(tokenlogin, os.Getenv(publickey))
		if !checkadmin {
			checkHR := IsHR(tokenlogin, os.Getenv(publickey))
			if !checkHR {
				resp.Status = false
				resp.Message = "Anda tidak bisa Insert data karena bukan HR atau admin"
			}
		} else {
			err := json.NewDecoder(r.Body).Decode(&req)
			if err != nil {
				resp.Message = "error parsing application/json: " + err.Error()
			} else {
				pass, err := pasproj.HashPass(req.Account.Password)
				if err != nil {
					resp.Status = false
					resp.Message = "Gagal Hash Code"
				}
				InsertDataEmployee(conn, colname, Employee{
					EmployeeId: req.EmployeeId,
					Name:       req.Name,
					Email:      req.Email,
					Phone:      req.Phone,
					Division: Division{
						DivId:   req.Division.DivId,
						DivName: req.Division.DivName,
					},
					Account: pasproj.User{
						Username: req.Account.Username,
						Password: pass,
						Role:     req.Account.Role,
					},
					Salary: Salary{
						BasicSalary:   req.Salary.BasicSalary,
						HonorDivision: req.Salary.HonorDivision,
					},
				})
				pasproj.InsertUserdata(conn, req.Account.Username, req.Account.Role, pass)
				resp.Status = true
				resp.Message = "Berhasil Insert data"
			}
		}
	}
	return pasproj.ReturnStringStruct(resp)
}

// Update data
func UpdateDataEmployees(MongoEnv, dbname, publickey string, r *http.Request) string {
	req := new(pasproj.Credential)
	resp := new(Employee)
	tokenlogin := r.Header.Get("Login")
	if tokenlogin == "" {
		req.Status = false
		req.Message = "Header Login Not Found"
	} else {
		err := json.NewDecoder(r.Body).Decode(&resp)
		if err != nil {
			req.Message = "error parsing application/json: " + err.Error()
		} else {
			checkadmin := IsAdmin(tokenlogin, os.Getenv(publickey))
			if !checkadmin {
				checkHR := IsHR(tokenlogin, os.Getenv(publickey))
				if !checkHR {
					req.Status = false
					req.Message = "Anda tidak bisa Insert data karena bukan HR atau admin"
				}
			} else {
				conn := pasproj.MongoCreateConnection(MongoEnv, dbname)
				UpdateEmployee(conn, context.Background(), Employee{
					EmployeeId: resp.EmployeeId,
					Name:       resp.Name,
					Email:      resp.Email,
					Phone:      resp.Phone,
					Division: Division{
						DivId:   resp.Division.DivId,
						DivName: resp.Division.DivName,
					},
					Account: pasproj.User{
						Username: resp.Account.Username,
						Password: resp.Account.Password,
						Role:     resp.Account.Role,
					},
					Salary: Salary{
						BasicSalary:   resp.Salary.BasicSalary,
						HonorDivision: resp.Salary.HonorDivision,
					},
				})
				req.Status = true
				req.Message = "Berhasil Update data"
			}
		}
	}
	return pasproj.ReturnStringStruct(req)
}

// Get One
func GetOneEmployee(PublicKey, MongoEnv, dbname, colname string, r *http.Request) string {
	req := new(ResponseEmployee)
	resp := new(RequestEmployee)
	conn := pasproj.MongoCreateConnection(MongoEnv, dbname)
	tokenlogin := r.Header.Get("Login")
	if tokenlogin == "" {
		req.Status = fiber.StatusBadRequest
		req.Message = "Header Login Not Found"
	} else {
		err := json.NewDecoder(r.Body).Decode(&resp)
		if err != nil {
			req.Message = "error parsing application/json: " + err.Error()
		} else {
			checkadmin := IsAdmin(tokenlogin, os.Getenv(PublicKey))
			if !checkadmin {
				checkHR := IsHR(tokenlogin, os.Getenv(PublicKey))
				if !checkHR {
					req.Status = fiber.StatusBadRequest
					req.Message = "Anda tidak bisa Get data karena bukan HR atau admin"
				}
			} else {
				datauser := GetOneEmployeeData(conn, colname, resp.EmployeeId)
				req.Status = fiber.StatusOK
				req.Message = "data User berhasil diambil"
				req.Data = datauser
			}
		}
	}
	return pasproj.ReturnStringStruct(req)
}

// GetAll
func GetAllEmployee(PublicKey, Mongoenv, dbname, colname string, r *http.Request) string {
	req := new(ResponseEmployeeBanyak)
	conn := pasproj.MongoCreateConnection(Mongoenv, dbname)
	tokenlogin := r.Header.Get("Login")
	if tokenlogin == "" {
		req.Status = fiber.StatusBadRequest
		req.Message = "Header Login Not Found"
	} else {
		checkadmin := IsAdmin(tokenlogin, os.Getenv(PublicKey))
		if !checkadmin {
			checkHR := IsHR(tokenlogin, os.Getenv(PublicKey))
			if !checkHR {
				req.Status = fiber.StatusBadRequest
				req.Message = "Anda tidak bisa Insert data karena bukan HR atau admin"
			}
		} else {
			datauser := GetAllEmployeeData(conn, colname)
			req.Status = fiber.StatusOK
			req.Message = "data User berhasil diambil"
			req.Data = datauser
		}
	}
	return pasproj.ReturnStringStruct(req)
}

// Delete Data
func DeleteEmployee(Mongoenv, publickey, dbname, colname string, r *http.Request) string {
	resp := new(Cred)
	req := new(RequestEmployee)
	conn := pasproj.MongoCreateConnection(Mongoenv, dbname)
	tokenlogin := r.Header.Get("Login")
	if tokenlogin == "" {
		resp.Status = fiber.StatusBadRequest
		resp.Message = "Token login tidak ada"
	} else {
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			resp.Message = "error parsing application/json: " + err.Error()
		} else {
			checkadmin := IsAdmin(tokenlogin, os.Getenv(publickey))
			if !checkadmin {
				resp.Status = fiber.StatusInternalServerError
				resp.Message = "kamu bukan admin"
			} else {
				_, err := DeleteEmployeeData(conn, colname, req.EmployeeId)
				if err != nil {
					resp.Status = fiber.StatusBadRequest
					resp.Message = "gagal hapus data"
				}
				resp.Status = fiber.StatusOK
				resp.Message = "data berhasil dihapus"
			}
		}
	}
	return pasproj.ReturnStringStruct(resp)
}