require "capybara"
require "capybara/rspec"
require_relative "../pages/applications"
require_relative "../pages/hosts"
require_relative "../pages/pools"
require_relative "../pages/service"
require_relative "../pages/servicesmap"
require_relative "../pages/user"

#
# Contains the list of page objects to be returned.  Login is only
# done if the session is invalid.
#
class UI
    include ::RSpec::Matchers
    include ::Capybara::DSL

    def initialize()
        @pages = {
            applications: Applications.new,
            hosts: Hosts.new,
            login: nil,
            pools: Pools.new,
            services: Service.new,
            servicesMap: ServicesMap.new,
            user: User.new,
        }
    end

    # Ensures that we've logged in.  Subsequent calls
    # will use the same page and not login again.
    def login()
        if self.LoginPage == nil || !verify_login?()
            puts "  \e[33m** Logging in to the UI **\e[0m"
            login_as(applicationUserID(), applicationPassword())
        end
    end

    def login_as(username, password)
        @pages[:login] = nil # Reset the login data.
        do_login(username, password)
    end

    def visit_login_page()
        @pages[:login] = Login.new if self.LoginPage == nil
        visit_login_page_impl(self.LoginPage)
    end

    def LoginPage
        return @pages[:login]
    end

    def HostsPage
        suppress_deploy_wizard()
        return @pages[:hosts]
    end

    def ApplicationsPage
        suppress_deploy_wizard()
        return @pages[:applications]
    end

    def ServicesMapPage
        suppress_deploy_wizard()
        return @pages[:servicesMap]
    end

    def PoolsPage
        suppress_deploy_wizard()
        return @pages[:pools]
    end

    def ServicesPage
        suppress_deploy_wizard()
        return @pages[:services]
    end

    def UserPage
        suppress_deploy_wizard()
        return @pages[:user]
    end

    # Look up the table data for the given port and remove it using
    # the UI.
    def remove_publicendpoint_port_json(name)
        remove_publicendpoint("table://ports/#{name}/PortAddr")
    end

    # Removes the public endpoint (vhost or port) by looking up the entry and
    # clicking the delete button.
    def remove_publicendpoint(name)
        name = getTableValue(name)
        self.ServicesPage.all(:xpath, "//table[@data-config='publicEndpointsTable']//tr").each do |tr|
            if tr.text.include?(name)
                btn = tr.find(:xpath, ".//button[@ng-click='clickRemovePublicEndpoint(publicEndpoint)']")
                if btn
                    btn.click
                    # confirm the removal
                    cnf = find(:xpath, "//div[@class='modal-content']//button", :text => "Remove and Restart Service")
                    cnf.click
                    refreshPage()
                    return true
                end
            end
        end
        return false
    end

    def click_add_publicendpoint_button()
        self.ServicesPage.addPublicEndpoint_button.click
        # wait till modal is done loading
        expect(self.ServicesPage).to have_no_css(".uilock", :visible => true)
    end

    def check_endpoint_unique_column?(ctitle, cvalue)
        found = 0
        self.ServicesPage.all(:xpath, "//table[@data-config='publicEndpointsTable']//tr//td[@data-title-text=#{ctitle}]").each do |td|
            if td.text.include?(cvalue)
                found += 1
            end
        end
        return true if found == 1

        return false
    end

    def check_endpoint_find?(c1, c2)
        self.ServicesPage.all(:xpath, "//table[@data-config='publicEndpointsTable']//tr").each do |tr|
            line=tr.text.upcase()
            if  line.include?(c1) && line.include?(c2)
                return true
            end
        end
        return false
    end


    def check_vhost_exists?(vhost)
        vhostName = getTableValue(vhost)
        searchStr = "https://#{vhostName}."
    
        found = false
        within(self.ServicesPage.publicEndpoints_table) do
            found = page.has_text?(searchStr)
        end
        return found
    end

    def check_public_port_exists?(port)
        portName = getTableValue(port)
        searchStr = ":#{portName}"

        found = false
        within(self.ServicesPage.publicEndpoints_table) do
            found = page.has_text?(searchStr)
        end
        return found
    end

    private

    ##
    # Login methods.
    #
    
    # Tries to access the pools page.  If we're redirected to the
    # login page, we need to login.
    def verify_login?()
        @pages[:pools].load # can't access the  property here since that tries to set cookies.
        return URI.parse(current_url).to_s.index("login") == nil
    end

    def do_login(username, password)
        login = Login.new

        visit_login_page_impl(login)
        suppress_deploy_wizard()
        fill_in_default_user_id(login, username)
        fill_in_default_password(login, password)
        click_signin_button(login)

        # Finally set the login variable
        @pages[:login] = login
    end

    def visit_login_page_impl(login)
        oldWait = setDefaultWaitTime(180)

        login.load
        expect(login).to be_displayed

        setDefaultWaitTime(oldWait)

        # wait till loading animation clears
        login.has_no_css?(".loading_wrapper")
    end

    def fill_in_default_user_id(login, username)
        login.userid_field.set username
    end

    def fill_in_default_password(login, password)
        login.password_field.set password
    end

    def click_signin_button(login)
        login.signin_button.click
    end

    def dump_cookies()
        if page.driver.browser.manage.all_cookies != nil
            page.driver.browser.manage.all_cookies.each do |cookie|
                printf "....cookie: %s\n",cookie[:name]
                return
            end
        else
            printf "....no cookies\n"
        end
    end

    def suppress_deploy_wizard()
        # dump_cookies()
        deploywizcookie = page.driver.browser.manage.cookie_named("autoRunWizardHasRun")
        if deploywizcookie != nil
            return
        end
        page.driver.browser.manage.add_cookie(
            {name:"autoRunWizardHasRun", value:"true"}
        )
    end
end